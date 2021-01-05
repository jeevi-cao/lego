package util

import (
	"errors"
	"net"
	"os"
)

//获取 本地 local ip
func GetLocalIp() (string, error) {
	//容器内获取方法
	var yidianLocalIP string
	yidianLocalIP = os.Getenv("YIDIAN_LOCAL_IP")
	ydIP := net.ParseIP(yidianLocalIP)
	if ydIP != nil {
		return yidianLocalIP, nil
	}
	//没有命中, 遍历网卡获取
	ifaces, err := net.Interfaces()
	// handle err
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		// handle err
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
