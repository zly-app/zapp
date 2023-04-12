package utils

import (
	"net"
)

// 获取实例名, 会返回本地ip, 如果无法获取返回默认值
func GetInstance(def string) string {
	ips := GetLocalIPs()
	if len(ips) == 0 {
		return def
	}
	return ips[0].String()
}

func GetLocalIPs() []net.IP {
	var ips []net.IP
	// 获取所有网卡接口信息
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, iface := range ifaces {
		// 排除本地回环接口和隧道接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}

		// 获取当前网卡接口的所有 IP 地址信息
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// 遍历当前网卡接口的所有 IP 地址
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// 排除本地 IP 地址和无效 IP 地址
			if ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() || ipNet.IP.IsLinkLocalMulticast() ||
				ipNet.IP.IsInterfaceLocalMulticast() || ipNet.IP.IsMulticast() || ipNet.IP.IsUnspecified() {
				continue
			}

			// 输出符合条件的 IP 地址
			ips = append(ips, ipNet.IP)
		}
	}
	return ips
}
