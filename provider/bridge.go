package provider

import (
	"net"
	"os"
	"os/exec"
)

func startBridge(listen string) (*os.Process, error) {
	cmd := exec.Command("./walletconnect-bridge", "-addr", listen)
	err := startProcess(cmd)
	return cmd.Process, err
}

func startProcess(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

func lanIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	r := "0.0.0.0"
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			if ipnet.IP[12] == 10 { // A
				ip := ipnet.IP.String()
				r = ip
			} else if ipnet.IP[12] == 172 && ipnet.IP[13] >= 16 && ipnet.IP[13] <= 31 { // B
				ip := ipnet.IP.String()
				r = ip
			} else if ipnet.IP[12] == 192 && ipnet.IP[13] == 168 { // C
				ip := ipnet.IP.String()
				return ip // use 192.168.x.x first
			}
		}
	}
	return r
}
