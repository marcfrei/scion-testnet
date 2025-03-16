package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/scionproto/scion/scion-pki/testcrypto"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "ifconfig":
		ifconfigCommand(os.Args[2:])
	case "cryptogen":
		cryptogenCommand(os.Args[2:])
	case "run":
		runCommand(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <subcommand> [arguments]\n\n", filepath.Base(os.Args[0]))
	fmt.Println("Available subcommands:")
	fmt.Println("  ifconfig [-c] <topology_path>  Configure networking for the topology")
	fmt.Println("  cryptogen [-c] <topology_path> Generate cryptographic material for the topology")
	fmt.Println("  run <topology_path>            Start SCION services using the specified topology")
	fmt.Println("  help                           Show this help message")
}

func ifconfigCommand(args []string) {
	if !checkPrivileges() {
		fmt.Println("Error: This command requires administrator/root privileges to modify network interfaces")
		os.Exit(1)
	}

	// Parse flags
	ifconfigFlags := flag.NewFlagSet("ifconfig", flag.ExitOnError)
	clean := ifconfigFlags.Bool("c", false, "clean (remove) IP addresses from interfaces")

	err := ifconfigFlags.Parse(args)
	if err != nil {
		fmt.Printf("Failed to parse ifconfig arguments: %v\n", err)
		os.Exit(1)
	}

	if ifconfigFlags.NArg() < 1 {
		fmt.Println("Usage: ifconfig [-c] <topology_path>")
		os.Exit(1)
	}

	topoPath := ifconfigFlags.Arg(0)
	networksFile := filepath.Join(topoPath, "networks.conf")

	if _, err := os.Stat(networksFile); os.IsNotExist(err) {
		fmt.Printf("Failed to open networks.conf\n")
		os.Exit(1)
	}

	ipAddrs := parseNetworksConfig(networksFile)

	var ipv4Count, ipv6Count int

	var ipv4Addresses []string
	var ipv6Addresses []string

	for ip := range ipAddrs {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			fmt.Printf("Warning: Invalid IP address format: %s\n", ip)
			continue
		}

		if parsedIP.To4() != nil {
			ipv4Addresses = append(ipv4Addresses, ip)
			ipv4Count++
		} else {
			ipv6Addresses = append(ipv6Addresses, ip)
			ipv6Count++
		}
	}

	if *clean {
		fmt.Printf("Removing %d IPv4 addresses and %d IPv6 addresses\n", ipv4Count, ipv6Count)
		switch runtime.GOOS {
		case "darwin":
			removeIPsDarwin(ipv4Addresses, ipv6Addresses)
		case "linux":
			removeIPsLinux(ipv4Addresses, ipv6Addresses)
		default:
			fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Adding %d IPv4 addresses and %d IPv6 addresses\n", ipv4Count, ipv6Count)
		switch runtime.GOOS {
		case "darwin":
			ifconfigDarwin(ipv4Addresses, ipv6Addresses)
		case "linux":
			ifconfigLinux(ipv4Addresses, ipv6Addresses)
		default:
			fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
			os.Exit(1)
		}
	}
}

func checkPrivileges() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	return string(output[:len(output)-1]) == "0"
}

func parseNetworksConfig(networksFile string) map[string]bool {
	file, err := os.Open(networksFile)
	if err != nil {
		fmt.Printf("Failed to open networks.conf: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	ipAddresses := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ip := strings.TrimSpace(parts[1])
		if ip != "" {
			ipAddresses[ip] = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading networks.conf: %v\n", err)
		os.Exit(1)
	}

	return ipAddresses
}

func ifconfigLinux(ipv4Addresses, ipv6Addresses []string) {
	for _, ip := range ipv4Addresses {
		cidr := ip
		if !strings.Contains(ip, "/") {
			cidr = ip + "/32"
		}

		cmd := exec.Command("ip", "addr", "add", cidr, "dev", "lo")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to add IPv4 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Added IPv4 address %s to loopback interface\n", ip)
		}
	}

	for _, ip := range ipv6Addresses {
		cidr := ip
		if !strings.Contains(ip, "/") {
			cidr = ip + "/128"
		}

		cmd := exec.Command("ip", "addr", "add", cidr, "dev", "lo")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to add IPv6 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Added IPv6 address %s to loopback interface\n", ip)
		}
	}
}

func removeIPsLinux(ipv4Addresses, ipv6Addresses []string) {
	for _, ip := range ipv4Addresses {
		cidr := ip
		if !strings.Contains(ip, "/") {
			cidr = ip + "/32"
		}

		cmd := exec.Command("ip", "addr", "del", cidr, "dev", "lo")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to remove IPv4 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Removed IPv4 address %s from loopback interface\n", ip)
		}
	}

	for _, ip := range ipv6Addresses {
		cidr := ip
		if !strings.Contains(ip, "/") {
			cidr = ip + "/128"
		}

		cmd := exec.Command("ip", "addr", "del", cidr, "dev", "lo")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to remove IPv6 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Removed IPv6 address %s from loopback interface\n", ip)
		}
	}
}

func ifconfigDarwin(ipv4Addresses, ipv6Addresses []string) {
	for _, ip := range ipv4Addresses {
		ipOnly := ip
		if strings.Contains(ip, "/") {
			ipOnly = strings.Split(ip, "/")[0]
		}

		cmd := exec.Command("ifconfig", "lo0", "alias", ipOnly)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to add IPv4 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Added IPv4 address %s to loopback interface\n", ip)
		}
	}

	for _, ip := range ipv6Addresses {
		ipOnly := ip
		prefixLen := "128"
		if strings.Contains(ip, "/") {
			parts := strings.Split(ip, "/")
			ipOnly = parts[0]
			if len(parts) > 1 {
				prefixLen = parts[1]
			}
		}

		cmd := exec.Command("ifconfig", "lo0", "inet6", ipOnly, "prefixlen", prefixLen, "alias")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to add IPv6 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Added IPv6 address %s to loopback interface\n", ip)
		}
	}
}

func removeIPsDarwin(ipv4Addresses, ipv6Addresses []string) {
	for _, ip := range ipv4Addresses {
		ipOnly := ip
		if strings.Contains(ip, "/") {
			ipOnly = strings.Split(ip, "/")[0]
		}

		cmd := exec.Command("ifconfig", "lo0", "-alias", ipOnly)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to remove IPv4 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Removed IPv4 address %s from loopback interface\n", ip)
		}
	}

	for _, ip := range ipv6Addresses {
		ipOnly := ip
		if strings.Contains(ip, "/") {
			ipOnly = strings.Split(ip, "/")[0]
		}

		cmd := exec.Command("ifconfig", "lo0", "inet6", ipOnly, "-alias")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to remove IPv6 address %s: %v\n", ip, err)
		} else {
			fmt.Printf("Removed IPv6 address %s from loopback interface\n", ip)
		}
	}
}

type commandPather string

func (s commandPather) CommandPath() string {
	return string(s)
}

func copyFile(src, dst string) {
	s, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()
	_, err = d.ReadFrom(s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func copyDir(src, dst string) {
	es, err := os.ReadDir(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, e := range es {
		n := e.Name()
		if n[0] != '.' {
			if e.IsDir() {
				panic("not yet implemented")
			} else if e.Type().IsRegular() {
				copyFile(filepath.Join(src, n), filepath.Join(dst, n))
			}
		}
	}
}

func collectCryptoPaths(cryptoPaths *[]string, dir string) {
	es, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, e := range es {
		n := e.Name()
		if n[0] != '.' {
			if e.IsDir() {
				p := filepath.Join(dir, n)
				if n == "certs" || n == "crypto" || n == "keys" || n == "trcs" {
					*cryptoPaths = append(*cryptoPaths, p)
				} else {
					collectCryptoPaths(cryptoPaths, p)
				}
			}
		}
	}
}

func genMasterKey(name string) {
	x := make([]byte, 16)
	n, err := rand.Read(x)
	if err != nil {
		panic(err)
	}
	if n != len(x) {
		panic("rand.Read failed")
	}
	f, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b := make([]byte, base64.StdEncoding.EncodedLen(len(x)))
	base64.StdEncoding.Encode(b, x)
	n, err = f.Write(b)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if n != len(b) {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cryptogenCommand(args []string) {
	cryptoFlags := flag.NewFlagSet("cryptogen", flag.ExitOnError)
	clean := cryptoFlags.Bool("c", false, "clean crypto directories")

	err := cryptoFlags.Parse(args)
	if err != nil {
		fmt.Printf("Failed to parse cryptogen arguments: %v\n", err)
		os.Exit(1)
	}

	if cryptoFlags.NArg() < 1 {
		fmt.Println("Usage: cryptogen [-c] <directory>")
		os.Exit(1)
	}

	genDir := cryptoFlags.Arg(0)
	trcDir := filepath.Join(genDir, "trcs")
	topoFile := filepath.Join(genDir, "topology.topo")

	var cryptoPaths []string

	collectCryptoPaths(&cryptoPaths, genDir)
	for _, p := range cryptoPaths {
		_ = os.RemoveAll(p)
	}
	cryptoPaths = nil

	if *clean {
		fmt.Printf("Cleaned crypto material in %s\n", genDir)
		return
	}

	cmd := testcrypto.Cmd(commandPather(""))
	cmd.SetArgs([]string{"-t", topoFile, "-o", genDir, "--as-validity", "28d"})
	stdout, stderr := os.Stdout, os.Stderr
	null, err := os.Open(os.DevNull)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	func() {
		os.Stdout, os.Stderr = null, null
		defer func() {
			os.Stdout, os.Stderr = stdout, stderr
		}()
		err = cmd.Execute()
	}()
	if err != nil {
		fmt.Printf("Failed to call command testcrypto: %v\n", err)
		os.Exit(1)
	}

	collectCryptoPaths(&cryptoPaths, genDir)
	for _, p := range cryptoPaths {
		_, q := filepath.Split(p)
		if q == "certs" {
			copyDir(trcDir, p)
		} else if q == "keys" {
			genMasterKey(filepath.Join(p, "master0.key"))
			genMasterKey(filepath.Join(p, "master1.key"))
		}
	}

	fmt.Printf("Generated crypto material in %s\n", genDir)
}

func runCommand(args []string) {
	scionPath := os.Getenv("SCION_PATH")
	if scionPath == "" {
		fmt.Printf("SCION_PATH environment variable is not set\n")
		os.Exit(1)
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	topoPath := args[0]

	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("Failed to create directory logs: %v", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("gen-cache", 0755); err != nil {
		fmt.Printf("Failed to create directory gen-cache: %v", err)
		os.Exit(1)
	}

	services := launchTopology(scionPath, topoPath)
	fmt.Printf("Ctrl+C to terminate\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println()

	for _, cmd := range services {
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("Failed to terminate service %d: %v\n", cmd.Process.Pid, err)
		}
		fmt.Printf("Terminated %s with PID %d\n", filepath.Base(cmd.Path), cmd.Process.Pid)
	}
}

func launchTopology(scionPath, topoPath string) []*exec.Cmd {
	var services []*exec.Cmd

	entries, err := os.ReadDir(topoPath)
	if err != nil {
		fmt.Printf("Failed to read directory %s: %v\n", topoPath, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(topoPath, entry.Name())
			svcs := processDirectory(scionPath, dirPath)
			services = append(services, svcs...)
		}
	}

	return services
}

func processDirectory(scionPath, dirPath string) []*exec.Cmd {
	var services []*exec.Cmd

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Failed to read directory %s: %v\n", dirPath, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
			filename := entry.Name()
			configPath := filepath.Join(dirPath, filename)

			var serviceType string
			if strings.HasPrefix(filename, "br") {
				serviceType = "router"
			} else if strings.HasPrefix(filename, "cs") {
				serviceType = "control"
			} else if strings.HasPrefix(filename, "sd") {
				serviceType = "daemon"
			} else if strings.HasPrefix(filename, "disp") {
				serviceType = "dispatcher"
			} else {
				// Skip files that don't match expected prefixes
				continue
			}

			logFilename := strings.TrimSuffix(filename, ".toml") + ".log"

			svc := startService(scionPath, serviceType, configPath, logFilename)
			if svc != nil {
				services = append(services, svc)
			}
		}
	}

	return services
}

func startService(scionPath, serviceType, configPath, logFilename string) *exec.Cmd {
	cmdPath := filepath.Join(scionPath, "bin", serviceType)

	cmd := exec.Command(cmdPath, "--config", configPath)

	logFile, err := os.OpenFile(filepath.Join("logs", logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start %s service with config %s : %v\n", serviceType, configPath, err)
		os.Exit(1)
	}

	fmt.Printf("Started %s with config %s (PID: %d)\n", serviceType, configPath, cmd.Process.Pid)

	return cmd
}
