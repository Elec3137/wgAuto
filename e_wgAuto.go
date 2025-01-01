package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

var wgName = "e_wgAuto"

var devgw string // Captured default gateway (for gateway reset functionality)

var wgconf_path = "/tmp/e_wgAuto.conf" // Path to configuration file automatically generated at each wg_start() execution

func set_default_gateway(endpoint string,addre string) {

	// Get current default gateway
	cmd := exec.Command("sh", "-c", "ip route show default | awk '/default/{print $3}'")
	out, err := cmd.Output()
	if err != nil {
		panic(err) // This must work without error, otherwise the script is (probably) functionless
	}
	devgw = strings.TrimSpace(string(out))

	// Set routes and modify resolv.conf
	exec.Command("ip", "route", "add", endpoint, "via", devgw).Run()
	exec.Command("ip", "route", "del", "default").Run()
	exec.Command("ip", "route", "add", "default", "via", addre).Run()
	exec.Command("sed", "-i", "s/nameserver .*/nameserver 8.8.8.8/g", "/etc/resolv.conf").Run()
}

func wg_start(conf string,endpoint string,addre string,nm string, default_gateway bool) {

	err := os.WriteFile(wgconf_path, []byte(conf), 0644)
	if err != nil {
		panic(err)
	}

	// WireGuard setup
	exec.Command("ip", "link", "add", "dev", wgName, "type", "wireguard").Run()
	exec.Command("ip", "address", "add", "dev", wgName, addre+nm).Run()
	exec.Command("ip", "link", "set", "up", "dev", wgName).Run()
	exec.Command("wg", "setconf", wgName, wgconf_path).Run()

	if default_gateway {
		set_default_gateway(endpoint,addre)
	}

}

func journalTrack(netnametoconnect string,conf string,endpoint string,addre string,nm string) {
	// Start following journalctl
	cmd := exec.Command("journalctl", "-f")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	var netname string

	// Read journal line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// For manual input into the logs with logger
		if strings.Contains(line, "wg_script_stop") {
			break
		}else if strings.Contains(line, "wg_start") {
			wg_start(conf,endpoint,addre,nm,true)
		}else if strings.Contains(line, "wg_stop") {
			exec.Command("ifconfig", wgName, "down").Run()
			if devgw!="" { // if variable is changed, set orginal default gateway
				exec.Command("ip", "route", "add", "default", "via", devgw).Run()
			}
		}else if strings.Contains(line, "wg_no_routing_start") {
			wg_start(conf,endpoint,addre,nm,false)
		}

		// Detects a line in the logs where the network name is listed and sets a variable to hold it
		if strings.Contains(line, "device (wlan0): Activation: starting connection") {
			fields := strings.Fields(line)
			if len(fields) >= 14 {
				netname = fields[12]
			}
		}

		// Check if netname matches
		if strings.Contains(line, "Activation: successful, device activated") && strings.Contains(line, "NetworkManager") && strings.Contains(line, "wlan") {
			if strings.Contains(netname, netnametoconnect) {
				wg_start(conf,endpoint,addre,nm,true)
			}
		}

		if (strings.Contains(line, "deactivating") && strings.Contains(line, "disconnected") && strings.Contains(line, "NetworkManager")) || (strings.Contains(line, "failed for connection") && strings.Contains(line, "NetworkManager")) {
			exec.Command("ifconfig", wgName, "down").Run()
			if devgw!="" { // (untested in this case, unessesary?)
				exec.Command("ip", "route", "add", "default", "via", devgw).Run()
			}
		}
	}
}


func main() {
	if strings.Contains(os.Args[1],"start") {
		conf := "[Interface]\nPrivateKey = " + os.Args[3] + "\n[Peer]\nPublicKey = " + os.Args[4] + "\nEndpoint = " + os.Args[5] + os.Args[6] + "\nAllowedIPs = 0.0.0.0/0 \nPersistentKeepalive = 10"
		journalTrack(os.Args[2],conf,os.Args[5],os.Args[7],os.Args[8])
	}else {
		println("Usage:  e_wgAuto start [netname] [privkey] [pubkey] [endpoint] [port] [addre] [nm]")
	}
}
