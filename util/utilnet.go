package util

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	CustomGlobalDNS = ""
)

func cleanDNSStr(dns string) (string, bool) {
	dns = strings.TrimSpace(dns)
	if dns != "" {
		if !strings.Contains(dns, ":") {
			dns += ":53"
		}
		return dns, true
	}
	return "", false
}

func InitDNSClient(client *http.Client, dns string) {
	if client == nil {
		return
	}
	if dns != CustomGlobalDNS {
		var ok bool
		dns, ok = cleanDNSStr(dns)
		if !ok {
			return
		}
	}
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := &net.Dialer{
					Timeout: 3 * time.Second,
				}
				return d.DialContext(ctx, "udp", dns)
			},
		},
	}
	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}
	client.Transport = transport
}

func InitCustomGlobalDNSClientIfSet(client *http.Client) {
	if UsingCustomGlobalDNS() {
		InitDNSClient(client, CustomGlobalDNS)
	}
}

func InitCustomGlobalDNSDefaultClientIfSet(dns string) {
	var ok bool
	dns, ok = cleanDNSStr(dns)
	if !ok {
		return
	}
	CustomGlobalDNS = dns
	InitDNSClient(http.DefaultClient, CustomGlobalDNS)
}

func UsingCustomGlobalDNS() bool {
	return CustomGlobalDNS != ""
}
