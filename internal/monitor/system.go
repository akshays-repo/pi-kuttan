package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
)

type Monitor struct{}

func New() *Monitor {
	return &Monitor{}
}

func (m *Monitor) GetSystemStats() (string, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return "", fmt.Errorf("error getting CPU usage: %w", err)
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("error getting memory usage: %w", err)
	}

	return fmt.Sprintf("CPU Usage: %.1f%%\nRAM Usage: %d MB / %d MB",
		cpuPercent[0],
		memInfo.Used/1024/1024,
		memInfo.Total/1024/1024), nil
}

func (m *Monitor) GetTemperature() (string, error) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return "", fmt.Errorf("error getting temperature: %w", err)
	}

	var cpuTemp float64
	for _, temp := range temps {
		if strings.Contains(strings.ToLower(temp.SensorKey), "cpu") {
			cpuTemp = temp.Temperature
			break
		}
	}

	return fmt.Sprintf("CPU Temperature: %.1f°C", cpuTemp), nil
}

func (m *Monitor) GetUptime() (string, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return "", fmt.Errorf("error getting uptime: %w", err)
	}

	duration := time.Duration(uptime) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return fmt.Sprintf("System Uptime: %d days, %d hours, %d minutes", 
		days, hours, minutes), nil
}

func (m *Monitor) GetDiskUsage() (string, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", fmt.Errorf("error getting disk partitions: %w", err)
	}

	var result strings.Builder
	result.WriteString("Disk Usage:\n")

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}
		result.WriteString(fmt.Sprintf("%s:\n  Used: %.1f GB / %.1f GB (%.1f%%)\n",
			partition.Mountpoint,
			float64(usage.Used)/1024/1024/1024,
			float64(usage.Total)/1024/1024/1024,
			usage.UsedPercent,
		))
	}

	return result.String(), nil
}
