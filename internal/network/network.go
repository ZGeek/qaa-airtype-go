package network

import (
	"net"
	"sort"
	"strings"
)

func GetHostIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func GetAllIPs() []string {
	var ips []string
	
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		ips = append(ips, "127.0.0.1")
		return ips
	}
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	
	if len(ips) == 0 {
		ips = append(ips, "127.0.0.1")
		return ips
	}
	
	priority192 := []string{}
	priority10 := []string{}
	other := []string{}
	virtual := []string{}
	
	for _, ip := range ips {
		if strings.HasPrefix(ip, "192.168.") {
			priority192 = append(priority192, ip)
		} else if strings.HasPrefix(ip, "10.") {
			priority10 = append(priority10, ip)
		} else if strings.HasPrefix(ip, "172.") {
			parts := strings.Split(ip, ".")
			if len(parts) >= 2 {
				second := parts[1]
				if second >= "16" && second <= "31" {
					virtual = append(virtual, ip)
					continue
				}
			}
			other = append(other, ip)
		} else if strings.HasPrefix(ip, "198.18.") {
			virtual = append(virtual, ip)
		} else {
			other = append(other, ip)
		}
	}
	
	sort.Strings(priority192)
	sort.Strings(priority10)
	sort.Strings(other)
	sort.Strings(virtual)
	
	result := append(priority192, priority10...)
	result = append(result, other...)
	result = append(result, virtual...)
	
	mainIP := GetHostIP()
	for i, ip := range result {
		if ip == mainIP {
			result = append(result[:i], result[i+1:]...)
			result = append([]string{mainIP}, result...)
			break
		}
	}
	
	return result
}