package util

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// 获取 MAC 地址
func getMacAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return "unknown"
}

// 获取 CPU ID
func getCpuID() string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("wmic", "cpu", "get", "ProcessorId")
	case "linux":
		cmd = exec.Command("cat", "/proc/cpuinfo")
	case "darwin":
		cmd = exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	default:
		return "unknown"
	}
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// 生成机器码
func generateMachineCode() string {
	data := getMacAddress() + getCpuID()
	//fmt.Println("data:", data)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
