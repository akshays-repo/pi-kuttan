package monitor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func (m *Monitor) GetTopProcesses() (string, error) {
	processes, err := process.Processes()
	if err != nil {
		return "", fmt.Errorf("error getting processes: %w", err)
	}

	type ProcessInfo struct {
		pid     int32
		name    string
		cpuPerc float64
		memPerc float32
	}

	var processInfos []ProcessInfo
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPerc, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memPerc, err := p.MemoryPercent()
		if err != nil {
			continue
		}

		processInfos = append(processInfos, ProcessInfo{
			pid:     p.Pid,
			name:    name,
			cpuPerc: cpuPerc,
			memPerc: memPerc,
		})
	}

	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].cpuPerc > processInfos[j].cpuPerc
	})

	var result strings.Builder
	result.WriteString("Top 5 Processes:\n")
	for i := 0; i < len(processInfos) && i < 5; i++ {
		result.WriteString(fmt.Sprintf("%d. %s (PID: %d)\n   CPU: %.1f%%, MEM: %.1f%%\n",
			i+1,
			processInfos[i].name,
			processInfos[i].pid,
			processInfos[i].cpuPerc,
			processInfos[i].memPerc,
		))
	}

	return result.String(), nil
}
