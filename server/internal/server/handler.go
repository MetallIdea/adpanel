package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetStatus(c *gin.Context) {
	status := Status{}

	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		status.CPU.UsagePercent = cpuPercent[0]
	}

	vm, err := mem.VirtualMemory()
	if err == nil {
		status.Memory.UsedMB = vm.Used / 1024 / 1024
		status.Memory.TotalMB = vm.Total / 1024 / 1024
		status.Memory.UsagePercent = vm.UsedPercent
	}

	diskUsage, err := disk.Usage("/")
	if err == nil {
		status.Disk.UsedMB = diskUsage.Used / 1024 / 1024
		status.Disk.TotalMB = diskUsage.Total / 1024 / 1024
		status.Disk.UsagePercent = diskUsage.UsedPercent
	}

	c.JSON(200, status)
}
