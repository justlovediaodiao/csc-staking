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
	localhost := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return localhost
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			// C
			if ipnet.IP[12] == 192 && ipnet.IP[13] == 168 {
				return ipnet.IP.String()
			}
			// A
			if ipnet.IP[12] == 10 {
				return ipnet.IP.String()
			}
			// B
			if ipnet.IP[12] == 172 && ipnet.IP[13] >= 16 && ipnet.IP[13] <= 31 {
				return ipnet.IP.String()
			}
		}
	}
	return localhost
}
