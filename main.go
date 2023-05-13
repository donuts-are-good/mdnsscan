package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"golang.org/x/crypto/ssh"
)

func main() {
	log.SetOutput(io.Discard)
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	doneCh := make(chan struct{})

	go func() {
		printEntries(entriesCh)
		close(doneCh)
	}()

	// Make a new mDNS lookup parameters structure
	params := &mdns.QueryParam{
		Service: "_services._dns-sd._udp",
		Domain:  "local",
		Timeout: time.Second * 5,
		Entries: entriesCh,
	}

	// Start the lookup
	lookupErr := mdns.Query(params)
	if lookupErr != nil && !strings.Contains(lookupErr.Error(), "Failed to query instance") {
		fmt.Println("Error from mDNS query:", lookupErr)
		return
	}

	// Close the entries channel after the query is done
	close(entriesCh)

	// Wait for the printEntries goroutine to finish
	<-doneCh

	// Wait for CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("\nLeaving...")
}

func printEntries(entriesCh chan *mdns.ServiceEntry) {
	totalDevices := 0
	totalOpenPorts := 0
	serviceCounts := make(map[string]int)

	for entry := range entriesCh {
		totalDevices++
		fmt.Println("Found entry:")
		fmt.Println("Name: ", entry.Name)
		fmt.Println("Host: ", entry.Host)
		fmt.Println("AddrV4: ", entry.AddrV4)
		fmt.Println("AddrV6: ", entry.AddrV6)
		fmt.Println("Port: ", entry.Port)
		if len(entry.InfoFields) > 0 {
			fmt.Println("Info:")
			for key, value := range entry.InfoFields {
				fmt.Printf("  %d: %s\n", key, value)
			}
		}
		fmt.Println("-------------------")
		fmt.Println("Scanning ports...")
		for port := 1; port <= 10000; port++ {
			fmt.Printf("Scanning port %d/10000...\r", port)
			address := fmt.Sprintf("%s:%d", entry.AddrV4, port)
			conn, err := net.DialTimeout("tcp", address, 3*time.Second)
			if err != nil {
				continue
			}
			conn.Close()
			totalOpenPorts++
			fmt.Printf("\nPort %d is open\n", port)
			service := identifyService(port)
			serviceMsg := probeService(address)
			if serviceMsg != "" {
				fmt.Printf("Service message: %s\n", serviceMsg)
			}
			if service != "" {
				fmt.Printf("Port %d is likely associated with %s\n", port, service)
				serviceCounts[service]++
				if service == "HTTP" {
					interactWithHTTP(entry.AddrV4.String(), port)
				} else if service == "SSH" {
					interactWithSSH(entry.AddrV4.String(), port)
				}
			}
		}
	}

	fmt.Println("Total devices:", totalDevices)
	fmt.Println("Total open ports:", totalOpenPorts)
	for service, count := range serviceCounts {
		fmt.Printf("Service %s found %d times\n", service, count)
	}
}

func identifyService(port int) string {
	switch port {
	case 22:
		return "SSH"
	case 80, 443:
		return "HTTP"
	case 554:
		return "RTSP"
	case 5353:
		return "mDNS"
	case 8008, 8009:
		return "Google Cast"
	case 9000:
		return "DLNA"
	case 123:
		return "NTP"
	case 1900:
		return "SSDP"
	case 2869:
		return "UPnP"
	case 5350, 5351:
		return "Bonjour Sleep Proxy"
	case 9090:
		return "AirPlay"
	case 9091:
		return "AirTunes"
	case 1901:
		return "PlayStation"
	case 32400:
		return "Plex Media Server"
	case 3689:
		return "iTunes"
	case 5357:
		return "Web Services for Devices"
	case 10243:
		return "Windows Remote Management"
	case 5000:
		return "Synology DiskStation Manager"
	case 5431:
		return "UPnP IGD"
	case 32469:
		return "Roku Media Server"
	default:
		return ""
	}
}

func interactWithHTTP(ip string, port int) {
	url := fmt.Sprintf("http://%s:%d", ip, port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making HTTP request to %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Got HTTP response from %s: %s\n", url, resp.Status)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading HTTP response body from %s: %v\n", url, err)
		return
	}
	bodyStr := string(bodyBytes)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100] + "..."
	}
	fmt.Printf("Beginning of response body from %s: %s\n", url, bodyStr)
}

func interactWithSSH(ip string, port int) {
	passwords := []string{"password", "admin", "administrator", "root", ""}

	for _, password := range passwords {
		sshConfig := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		address := fmt.Sprintf("%s:%d", ip, port)
		client, err := ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			fmt.Printf("Error connecting to SSH server at %s: %v\n", address, err)
			continue
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			fmt.Printf("Error creating SSH session: %v\n", err)
			continue
		}
		defer session.Close()

		// Attempt to execute a command to check if the connection responds to root
		cmd := "whoami"
		output, err := session.Output(cmd)
		if err != nil {
			fmt.Printf("Failed to execute command: %v\n", err)
			continue
		}

		// Check if the output contains "root"
		if strings.TrimSpace(string(output)) == "root" {
			fmt.Printf("SSH server responds to root user with password: %s\n", password)
		} else {
			fmt.Printf("SSH server does not respond to root user with password: %s\n", password)
		}

		fmt.Printf("Connected to SSH server at %s\n", address)
	}
}

func probeService(address string) string {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()

	// Set a deadline for the read operation
	deadline := time.Now().Add(2 * time.Second)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		fmt.Println("Failed to set read deadline:", err)
		return ""
	}

	// Read the first 1024 bytes from the connection
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return ""
	}

	// Convert to a string and return
	return string(buf[:n])
}
