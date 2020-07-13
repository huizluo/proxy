package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
)

func PathExists(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
//Keygen gen ssl key and crt
func Keygen() (err error) {
	cmd := exec.Command("sh", "-c", "openssl genrsa -out proxy.key 2048")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("err:%s", err)
		return
	}
	fmt.Println(string(out))
	cmd = exec.Command("sh", "-c", `openssl req -new -key proxy.key -x509 -days 3650 -out proxy.crt -subj /C=CN/ST=BJ/O="Localhost Ltd"/CN=proxy`)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("err:%s", err)
		return
	}
	fmt.Println(string(out))
	return
}
//GetAllInterfaceAddr
func GetAllInterfaceAddr() ([]net.IP, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	addresses := []net.IP{}
	for _, iface := range ifaces {

		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		// if iface.Flags&net.FlagLoopback != 0 {
		// 	continue // loopback interface
		// }
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// if ip == nil || ip.IsLoopback() {
			// 	continue
			// }
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			addresses = append(addresses, ip)
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no address Found, net.InterfaceAddrs: %v", addresses)
	}
	//only need first
	return addresses, nil
}
