package common

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Color
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"

	// DNS config
	DNS_CONFIG_FILE         = ".config/403unlocker/dns.conf"
	CHECKED_DNS_CONFIG_FILE = ".config/403unlocker/checked_dns.conf"
	DOCKER_CONFIG_FILE      = ".config/403unlocker/dockerRegistry.conf"
	DNS_CONFIG_URL          = "https://raw.githubusercontent.com/403unlocker/403Unlocker-cli/refs/heads/main/config/dns.conf"
	DOCKER_CONFIG_URL       = "https://raw.githubusercontent.com/403unlocker/403Unlocker-cli/refs/heads/main/config/dockerRegistry.conf"
)

// FormatDataSize converts the size in bytes to a human-readable string in KB, MB, or GB.
func FormatDataSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d Bytes", bytes)
	}
}

func DownloadConfigFile(url, path string) error {

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		fmt.Println("HOME environment variable not set")
		os.Exit(1)
	}
	filePath := homeDir + "/" + path

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return err
	}

	out, err := os.Create(filePath)

	if err != nil {
		fmt.Println(err)

		return err
	}
	defer out.Close()

	if err != nil {
		fmt.Println("Could not download config file.")
		return err
	}

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Could not get the response: ", err)
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		fmt.Println("Could not copy content file")
		return err
	}

	return nil
}

func WriteDNSToFile(filename string, dnsList []string) error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		fmt.Println("HOME environment variable not set")
		os.Exit(1)
	}

	filename = homeDir + "/" + filename

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filename, err)
			return err
		}
		file.Close()
	}

	content := strings.Join(dnsList, " ")

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filename, err)
		return err
	}

	return nil
}

func ReadDNSFromFile(filename string) ([]string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		fmt.Println("HOME environment variable not set")
		os.Exit(1)
	}
	filename = homeDir + "/" + filename
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	dnsServers := strings.Fields(string(data))
	return dnsServers, nil
}

func ChangeDNS(dns string) *http.Client {
	dialer := &net.Dialer{}
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dnsServer := fmt.Sprintf("%s:53", dns)
			return dialer.DialContext(ctx, "udp", dnsServer)
		},
	}
	customDialer := &net.Dialer{
		Resolver: customResolver,
	}
	transport := &http.Transport{
		DialContext: customDialer.DialContext,
	}
	client := &http.Client{
		Transport: transport,
	}
	return client
}
