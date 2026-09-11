package sysmon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	goruntime "runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SysStats struct {
	CPU     string `json:"cpu"`
	RAM     string `json:"ram"`
	Swap    string `json:"swap"`
	Disk    string `json:"disk"`
	NetDL   string `json:"netDl"`
	NetUL   string `json:"netUl"`
	GameCPU string `json:"gameCpu"`
	GameRAM string `json:"gameRam"`
}

type GameTelemetryData struct {
	Timestamp int64   `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
}

var (
	mu             sync.Mutex
	activeGamePID  int32
	activeGameID   string
	rollingBuffer  []GameTelemetryData
	maxBufferLines = 450 // 15 mins at 2s interval
	sessionStart   time.Time
)

// PlaytimeData stores per-game and total playtime in seconds
type PlaytimeData struct {
	Total   int64            `json:"total"`
	PerGame map[string]int64 `json:"per_game"`
}

func SetActiveGame(pid int32, gameID string) {
	mu.Lock()
	defer mu.Unlock()
	activeGamePID = pid
	activeGameID = gameID
	rollingBuffer = make([]GameTelemetryData, 0, maxBufferLines)
	sessionStart = time.Now()
}

func UpdateActiveGamePID(pid int32) {
	mu.Lock()
	defer mu.Unlock()
	activeGamePID = pid
}

func ClearActiveGameAndSave(logsDir string) {
	mu.Lock()
	defer mu.Unlock()
	if activeGameID == "" {
		activeGamePID = 0
		return
	}

	// Save to post-mortem
	os.MkdirAll(logsDir, 0755)
	postMortemPath := filepath.Join(logsDir, activeGameID+"_postmortem.json")
	data, _ := json.MarshalIndent(rollingBuffer, "", "  ")
	os.WriteFile(postMortemPath, data, 0644)

	// Update playtime
	duration := time.Since(sessionStart).Seconds()
	if duration > 0 {
		home, _ := os.UserHomeDir()
		dataDir := filepath.Join(home, ".rift", "data")
		os.MkdirAll(dataDir, 0755)
		playtimeFile := filepath.Join(dataDir, "playtime.json")

		var pt PlaytimeData
		pt.PerGame = make(map[string]int64)
		if b, err := os.ReadFile(playtimeFile); err == nil {
			json.Unmarshal(b, &pt)
		}
		if pt.PerGame == nil {
			pt.PerGame = make(map[string]int64)
		}

		added := int64(duration)
		pt.Total += added
		pt.PerGame[activeGameID] += added

		if out, err := json.MarshalIndent(pt, "", "  "); err == nil {
			os.WriteFile(playtimeFile, out, 0644)
		}
	}

	activeGamePID = 0
	activeGameID = ""
	rollingBuffer = nil
}

func IsGameRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return activeGamePID != 0 || activeGameID != ""
}

// GetLastSessionTelemetry loads the most recent postmortem file
func GetLastSessionTelemetry(logsDir string) string {
	mu.Lock()
	defer mu.Unlock()

	// Find the most recently modified _postmortem.json file
	var newestFile string
	var newestTime time.Time

	files, _ := os.ReadDir(logsDir)
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" && len(f.Name()) > 16 && f.Name()[len(f.Name())-16:] == "_postmortem.json" {
			info, err := f.Info()
			if err == nil && info.ModTime().After(newestTime) {
				newestTime = info.ModTime()
				newestFile = filepath.Join(logsDir, f.Name())
			}
		}
	}

	if newestFile != "" {
		data, err := os.ReadFile(newestFile)
		if err == nil {
			return string(data)
		}
	}
	return "[]"
}

func GetPlaytimeStats() string {
	home, _ := os.UserHomeDir()
	playtimeFile := filepath.Join(home, ".rift", "data", "playtime.json")
	data, err := os.ReadFile(playtimeFile)
	if err == nil {
		return string(data)
	}
	return `{"total": 0, "per_game": {}}`
}

func StartTelemetry(ctx context.Context) {
	var prevNet net.IOCountersStat
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		prevNet = netStats[0]
	}
	prevTime := time.Now()

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := SysStats{
					CPU:     "--%",
					RAM:     "-- GB",
					Swap:    "-- MB",
					Disk:    "-- GB Free",
					NetDL:   "-- MB/s",
					NetUL:   "-- MB/s",
					GameCPU: "--%",
					GameRAM: "-- MB",
				}

				if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
					stats.CPU = fmt.Sprintf("%.1f%%", p[0])
				}

				if v, err := mem.VirtualMemory(); err == nil {
					usedGB := float64(v.Used) / 1024 / 1024 / 1024
					totalGB := float64(v.Total) / 1024 / 1024 / 1024
					stats.RAM = fmt.Sprintf("%.1f / %.1f GB", usedGB, totalGB)
				}
				if s, err := mem.SwapMemory(); err == nil {
					usedMB := float64(s.Used) / 1024 / 1024
					stats.Swap = fmt.Sprintf("%.1f MB", usedMB)
				}
				if d, err := disk.Usage("/"); err == nil {
					freeGB := float64(d.Free) / 1024 / 1024 / 1024
					stats.Disk = fmt.Sprintf("%.1f GB Free", freeGB)
				}

				currNetStats, err := net.IOCounters(false)
				currTime := time.Now()
				if err == nil && len(currNetStats) > 0 {
					currNet := currNetStats[0]
					duration := currTime.Sub(prevTime).Seconds()
					if duration > 0 {
						bytesRecv := currNet.BytesRecv - prevNet.BytesRecv
						bytesSent := currNet.BytesSent - prevNet.BytesSent
						dlMBps := (float64(bytesRecv) / duration) / 1024 / 1024
						ulMBps := (float64(bytesSent) / duration) / 1024 / 1024
						stats.NetDL = fmt.Sprintf("%.2f MB/s", dlMBps)
						stats.NetUL = fmt.Sprintf("%.2f MB/s", ulMBps)
					}
					prevNet = currNet
					prevTime = currTime
				}

				mu.Lock()
				if activeGamePID > 0 {
					if exists, _ := process.PidExists(activeGamePID); exists {
						cpuNorm, gameRAM := getFullProcessTreeMetrics(activeGamePID)
						stats.GameCPU = fmt.Sprintf("%.1f%%", cpuNorm)
						stats.GameRAM = fmt.Sprintf("%.1f MB", gameRAM)

						rollingBuffer = append(rollingBuffer, GameTelemetryData{
							Timestamp: time.Now().Unix(),
							CPU:       cpuNorm,
							RAM:       gameRAM,
						})
						if len(rollingBuffer) > maxBufferLines {
							rollingBuffer = rollingBuffer[1:]
						}
					} else {
						// Process died or not accessible
						activeGamePID = 0
					}
				}
				mu.Unlock()

				runtime.EventsEmit(ctx, "sys-stats", stats)
			}
		}
	}()
}

func getFullProcessTreeMetrics(pid int32) (float64, float64) {
	root, err := process.NewProcess(pid)
	if err != nil {
		return 0, 0
	}

	procs := []*process.Process{root}
	if children, err := root.Children(); err == nil {
		procs = append(procs, children...)
		for _, ch := range children {
			if grandchildren, err := ch.Children(); err == nil {
				procs = append(procs, grandchildren...)
			}
		}
	}

	totalCPU := 0.0
	totalRAM := 0.0
	for _, p := range procs {
		if c, err := p.CPUPercent(); err == nil {
			totalCPU += c
		}
		if mem, err := p.MemoryInfo(); err == nil && mem != nil {
			totalRAM += float64(mem.RSS) / 1024 / 1024
		}
	}

	// Normalize CPU across all logical cores so it represents true 0% to 100% total system CPU
	numCores := float64(goruntime.NumCPU())
	if numCores <= 0 {
		numCores = 1
	}
	normalizedCPU := totalCPU / numCores
	if normalizedCPU > 100.0 {
		normalizedCPU = 100.0
	}

	return normalizedCPU, totalRAM
}
