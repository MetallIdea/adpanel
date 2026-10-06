package server

type Status struct {
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   DiskInfo   `json:"disk"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	UsedMB   uint64  `json:"used_mb"`
	TotalMB  uint64  `json:"total_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskInfo struct {
	UsedMB     uint64  `json:"used_mb"`
	TotalMB    uint64  `json:"total_mb"`
	UsagePercent float64 `json:"usage_percent"`
}
