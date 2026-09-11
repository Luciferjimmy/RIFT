package optimizer

import (
	"fmt"
	"regexp"
	"strings"
)

// SteamAppIDPattern matches common Steam AppID representations.
var steamAppIDPattern = regexp.MustCompile(`^\d{1,10}$`)

// IdentityResolver resolves a raw game identifier (name, Steam AppID, etc.)
// into a canonical GameIdentity used for all downstream lookups.
type IdentityResolver struct {
	db *DBProvider
}

// NewIdentityResolver creates a resolver backed by the given DB provider.
func NewIdentityResolver(db *DBProvider) *IdentityResolver {
	return &IdentityResolver{db: db}
}

// Resolve attempts to build a GameIdentity from the provided hints.
// At least one of steamID or name should be non-empty.
func (r *IdentityResolver) Resolve(steamID, name string) (*GameIdentity, error) {
	ident := &GameIdentity{
		Name:       name,
		SteamAppID: steamID,
	}

	// Validate the Steam AppID if provided.
	if steamID != "" && !steamAppIDPattern.MatchString(steamID) {
		return nil, fmt.Errorf("invalid Steam AppID format: %q", steamID)
	}

	// If we only have a name, try to find a Steam AppID from PCGamingWiki.
	if ident.SteamAppID == "" && ident.Name != "" {
		pcgw, err := r.db.QueryPCGamingWiki("", ident.Name)
		if err == nil && pcgw != nil {
			ident.SteamAppID = pcgw.SteamAppID
			if ident.Name == "" {
				ident.Name = pcgw.PageName
			}
		}
	}

	// If we still have no name, derive it from SteamID via ProtonDB.
	if ident.Name == "" && ident.SteamAppID != "" {
		entries, err := r.db.QueryProtonDB(ident.SteamAppID)
		if err == nil && len(entries) > 0 {
			ident.Name = entries[0].GameName
		}
	}

	// Final fallback: use the input as-is.
	if ident.Name == "" {
		ident.Name = name
	}

	// Build alias list for fuzzy matching.
	ident.Aliases = buildAliases(ident.Name)
	return ident, nil
}

// buildAliases generates a set of alternative names for fuzzy matching.
func buildAliases(name string) []string {
	seen := map[string]bool{name: true}
	var aliases []string

	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			aliases = append(aliases, s)
		}
	}

	// Lowercase variant.
	add(strings.ToLower(name))

	// Strip common suffixes like "™", "®", "- Windows", "(Windows)".
	cleaned := name
	cleaned = strings.TrimSuffix(cleaned, "™")
	cleaned = strings.TrimSuffix(cleaned, "®")
	cleaned = strings.TrimSuffix(cleaned, "(Windows)")
	cleaned = strings.TrimSuffix(cleaned, "- Windows")
	cleaned = strings.TrimSpace(cleaned)
	add(cleaned)

	// Substring before colon (many games have "Title: Subtitle" format).
	if idx := strings.Index(name, ":"); idx > 0 {
		add(strings.TrimSpace(name[:idx]))
	}

	return aliases
}
