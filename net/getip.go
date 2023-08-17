package net

import (
	"fmt"
	"net"
)

// getIps
// 提取网卡ip
func GetIps() (ips string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
		return ips
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						//fmt.Println(ipnet.IP.String())
						//ips = append(ips, ipnet.IP.String())
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ips
}
