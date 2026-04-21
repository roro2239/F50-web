package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	go func() {
		// 第一次采样
		stats1, err := readCPUStats()
		if err != nil {
			print("CPU占用率获取失败", err)
			return
		}
		//print(stats1)
		for {

			time.Sleep(1 * time.Second)

			// 第二次采样
			stats2, err := readCPUStats()
			if err != nil {
				continue
			}

			// 计算总时间差
			totalDelta := totalCPU(stats2) - totalCPU(stats1)
			idleDelta := stats2.Idle - stats1.Idle

			// 计算使用率百分比
			usage := (1 - float64(idleDelta)/float64(totalDelta)) * 100
			CPUusage = usage

			stats1 = stats2

			MEMusage, _ = getMemoryUsage()

			MAXtemp, _ = getMaxTemperature()

			Battery, BatteryCharging = getBattery()
		}
	}()

}

type CPUStats struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
}

func getCPUUsage() (float64, error) {
	// 第一次采样
	stats1, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)

	// 第二次采样
	stats2, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	// 计算总时间差
	totalDelta := totalCPU(stats2) - totalCPU(stats1)
	idleDelta := stats2.Idle - stats1.Idle

	if totalDelta == 0 {
		return 0, nil
	}

	// 计算使用率百分比
	usage := (1 - float64(idleDelta)/float64(totalDelta)) * 100
	return usage, nil
}

func readCPUStats() (CPUStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return CPUStats{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			parts := strings.Fields(line)[1:]
			var stats CPUStats
			values := make([]uint64, len(parts))
			for i, part := range parts {
				val, err := strconv.ParseUint(part, 10, 64)
				if err != nil {
					return CPUStats{}, err
				}
				values[i] = val
			}
			stats.User = values[0]
			stats.Nice = values[1]
			stats.System = values[2]
			stats.Idle = values[3]
			stats.IOWait = values[4]
			stats.IRQ = values[5]
			stats.SoftIRQ = values[6]
			stats.Steal = values[7]
			return stats, nil
		}
	}
	return CPUStats{}, fmt.Errorf("cpu stats not found")
}

func totalCPU(s CPUStats) uint64 {
	return s.User + s.Nice + s.System + s.Idle + s.IOWait + s.IRQ + s.SoftIRQ + s.Steal
}

func getMemoryUsage() (float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			memTotal = parseMemLine(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			memAvailable = parseMemLine(line)
		}
	}

	if memTotal == 0 {
		return 0, fmt.Errorf("MemTotal not found")
	}

	// 计算已用内存 (MemTotal - MemAvailable)
	used := memTotal - memAvailable
	usage := (float64(used) / float64(memTotal)) * 100
	return usage, nil
}

func parseMemLine(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	val, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0
	}
	return val * 1024 // 转换为字节
}

func sys_main() {
	total, used, avail, usage := getDiskUsage("/sdcard")
	print(total, used, avail, usage)
	print(total/1e9, used/1e9, avail/1e9, usage)
	for {
		printf("CPU: %.1f %%\n", CPUusage)
		printf("MEM: %.1f %%\n", MEMusage)
		printf("温度: %.1f °C\n", MAXtemp)
		time.Sleep(time.Second)
	}

}

func getMaxTemperature() (float64, error) {
	thermalPath := "/sys/class/thermal"
	zones, err := ioutil.ReadDir(thermalPath)
	if err != nil {
		return 0, fmt.Errorf("无法访问温度传感器目录: %v", err)
	}

	maxTemp := -math.MaxFloat64 // 初始化为最小可能值
	foundValid := false

	for _, zone := range zones {
		zoneName := zone.Name()
		if !strings.HasPrefix(zoneName, "thermal_zone") {
			continue
		}

		// 读取温度值
		tempPath := filepath.Join(thermalPath, zoneName, "temp")
		tempData, err := ioutil.ReadFile(tempPath)
		if err != nil {
			continue // 跳过无法读取的传感器
		}

		// 转换温度值
		tempStr := strings.TrimSpace(string(tempData))
		tempMilli, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			continue // 跳过无效数据
		}

		// 转换为摄氏度（多数系统使用千分位单位）
		currentTemp := tempMilli / 1000
		if currentTemp > maxTemp {
			maxTemp = currentTemp
			foundValid = true
		}
	}

	if !foundValid {
		return 0, fmt.Errorf("未找到有效的温度传感器")
	}

	return maxTemp, nil
}

// 返回 总空间（字节） 已用空间（字节） 可用空间（字节） 使用率百分比
func getDiskUsage(mountPoint string) (uint64, uint64, uint64, float64) {
	return 0, 0, 0, 0
	// var stat syscall.Statfs_t

	// err := syscall.Statfs(mountPoint, &stat)
	// if err != nil {
	// 	return 0, 0, 0, 0
	// }

	// // 计算各空间值
	// total := stat.Blocks * uint64(stat.Bsize)
	// avail := stat.Bavail * uint64(stat.Bsize)
	// used := total - avail

	// usage := 0.0
	// if total > 0 {
	// 	usage = (float64(used) / float64(total)) * 100
	// }

	// return total, used, avail, usage
}

func getBattery() (int, bool) {
	a := NewCommandExecutor("cat /sys/class/power_supply/battery/capacity")
	e := a.Start()
	if e != nil {
		return -1, false
	}
	a.Wait()
	r := strings.Join(a.GetLogs(), "")
	r = strings.ReplaceAll(r, "\n", "")
	ir, e := strconv.Atoi(r)
	if e != nil {
		return -1, false
	}

	a = NewCommandExecutor("cat /sys/class/power_supply/battery/status")
	e = a.Start()
	if e != nil {
		return ir, false
	}
	a.Wait()
	r = strings.Join(a.GetLogs(), "")
	r = strings.ReplaceAll(r, "\n", "")
	if r == "Charging" {
		return ir, true
	}
	return ir, false
}
