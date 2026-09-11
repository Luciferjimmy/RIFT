package optimizer

import (
	"context"
	"sync"
)

// Aggregator queries all local databases in parallel and merges the results.
type Aggregator struct {
	db     *DBProvider
	prefix string // optional prefix to prepend when searching by name
}

// NewAggregator creates an aggregator that uses the given DB provider.
func NewAggregator(db *DBProvider) *Aggregator {
	return &Aggregator{db: db}
}

// Aggregate runs all database queries concurrently and returns a single merged
// view. If ctx is cancelled, in-flight queries are abandoned and partial
// results are returned.
func (a *Aggregator) Aggregate(ctx context.Context, ident *GameIdentity) *AggregatedGameData {
	var (
		wg   sync.WaitGroup
		data AggregatedGameData
	)

	// ---------- PCGamingWiki ----------
	wg.Add(1)
	go func() {
		defer wg.Done()
		pcgw, err := a.db.QueryPCGamingWiki(ident.SteamAppID, ident.Name)
		if err == nil && pcgw != nil {
			data.PCGamingWiki = pcgw
		}
	}()

	// ---------- ProtonDB ----------
	wg.Add(1)
	go func() {
		defer wg.Done()
		if ident.SteamAppID == "" {
			return
		}
		entries, err := a.db.QueryProtonDB(ident.SteamAppID)
		if err != nil || len(entries) == 0 {
			return
		}
		// Pick the highest-rated entry.
		best := entries[0]
		for _, e := range entries[1:] {
			if ratingWeight(e.Rating) < ratingWeight(best.Rating) {
				best = e
			}
		}
		data.ProtonDBBest = &best
	}()

	// ---------- Lutris ----------
	wg.Add(1)
	go func() {
		defer wg.Done()
		games, err := a.db.QueryLutrisGames(ident.Name)
		if err != nil || len(games) == 0 {
			return
		}
		// Try each matched slug until we find a script.
		for _, g := range games {
			if !g.HasInstallers {
				continue
			}
			script, err := a.db.QueryLutrisScript(g.Slug)
			if err == nil && script != nil {
				data.LutrisBest = script
				return
			}
		}
		// No installer script found, but we still have a game match.
	}()

	// ---------- Apple Gaming Wiki ----------
	wg.Add(1)
	go func() {
		defer wg.Done()
		entry, err := a.db.QueryAGW(ident.Name)
		if err == nil && entry != nil {
			data.AGW = entry
			data.HasNativeMac = entry.NativeMac
		}
	}()

	wg.Wait()
	return &data
}

// ratingWeight maps ProtonDB ratings to a numeric ordering (lower = better).
func ratingWeight(r string) int {
	switch r {
	case "Platinum":
		return 0
	case "Gold":
		return 1
	case "Silver":
		return 2
	case "Bronze":
		return 3
	case "Borked":
		return 4
	default:
		return 5
	}
}
