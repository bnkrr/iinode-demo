package main

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// https://cloudflare.com/cdn-cgi/trace
// https://one.one.one.one/cdn-cgi/trace
// https://1.0.0.1/cdn-cgi/trace
// https://cloudflare-dns.com/cdn-cgi/trace
// https://cloudflare-eth.com/cdn-cgi/trace
// https://cloudflare-ipfs.com/cdn-cgi/trace
// https://workers.dev/cdn-cgi/trace
// https://pages.dev/cdn-cgi/trace
// https://cloudflare.tv/cdn-cgi/trace
// https://icanhazip.com/cdn-cgi/trace

func GetIPAndLocationFromCloudflare(url string, ipVerion int) (string, string, error) {
	var client *http.Client
	if ipVerion == 4 {
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("tcp4", addr)
			},
		}
		client = &http.Client{
			Transport: transport,
			Timeout:   3 * time.Second,
		}
	} else if ipVerion == 6 {
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("tcp6", addr)
			},
		}
		client = &http.Client{
			Transport: transport,
			Timeout:   3 * time.Second,
		}
	} else {
		client = &http.Client{}
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	traceData := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			traceData[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return traceData["ip"], traceData["loc"], err
}

func dbusId() (string, error) {
	id, err := os.ReadFile("/var/lib/dbus/machine-id")
	if err != nil {
		id, err = os.ReadFile("/etc/machine-id")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Trim(string(id), "\n")), nil
}

type NodeMetadata struct {
	PublicIp string
	Id       string
	Location string
}

func (n *NodeMetadata) GenerateId() {
	id, err := dbusId()
	if err != nil {
		id = "default"
	}
	n.Id = id
}

func (n *NodeMetadata) GetPublicIp() {
	ip, loc, err := GetIPAndLocationFromCloudflare("https://cloudflare-eth.com/cdn-cgi/trace", 4)
	if err != nil {
		n.PublicIp = "Unknown"
		n.Location = "Unknown"
	} else {
		n.PublicIp = ip
		n.Location = loc
	}
}

func (n *NodeMetadata) Refresh() {
	n.GenerateId()
	n.GetPublicIp()
}
