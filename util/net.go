package util


import (
	"errors"
	"net"
	"net/http"
	"strings"
)

/*
 * net util tool
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Net struct {
}

//get local ip
func (f *Net) GetLocalIp() ([]string, error) {
	addrSlice, err := net.InterfaceAddrs()
	if err != nil || addrSlice == nil {
		return nil, err
	}
	ips := make([]string, 0)
	for _, v := range addrSlice {
		if ipNet, ok := v.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.To4().String())
			}
		}
	}
	return ips, nil
}

//get out ip
func (f *Net) GetOutIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil || conn == nil {
		return "", err
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr == nil {
		return "", errors.New("can't get udp address")
	}
	return addr.IP.String(), nil
}

//gather all ip from client
func (f *Net) GetClientAllIp(r *http.Request) []string {
	var tempStr string
	var ipSlice = make([]string, 0)

	//get original data
	clientAddress := r.RemoteAddr
	xRealIp := r.Header.Get("X-Real-IP")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	//analyze general ip
	if clientAddress != "" {
		tempStr = f.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-real-ip
	if xRealIp != "" {
		tempStr = f.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-forward-for
	//like:192.168.0.1,192.168.0.2
	if xForwardedFor != "" {
		tempSlice := strings.Split(xForwardedFor, ",")
		if len(tempSlice) > 0 {
			for _, tmpAddr := range tempSlice {
				tempStr = f.analyzeClientIp(tmpAddr)
				if tempStr != "" {
					ipSlice = append(ipSlice, tempStr)
				}
			}
		}
	}

	return ipSlice
}

//analyze client ip
func (f *Net) analyzeClientIp(address string) string {
	tempSlice := strings.Split(address, ":")
	if len(tempSlice) < 1 {
		return ""
	}
	return tempSlice[0]
}