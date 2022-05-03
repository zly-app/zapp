package utils

import (
	"net"
)

// 获取实例名, 会返回本地ip, 如果无法获取返回默认值
func GetInstance(def string) string {
	var ips []string
	address, _ := net.InterfaceAddrs()

	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	if len(ips) > 0 {
		return ips[0]
	}
	return def
}
