package client

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"net"
	"strings"

	randv2 "math/rand/v2"
)

var ErrInvalidHostAlias = errors.New("invalid host alias")

func ParseHostAlias(x string) (*HostAlias, error) {
	if x == "" {
		return nil, fmt.Errorf("empty host alias: %w", ErrInvalidHostAlias)
	}

	ha := &HostAlias{}

	if strings.IndexByte(x, '[') == 0 {
		// ipv6
		end := strings.IndexByte(x, ']')
		if end == -1 {
			return nil, fmt.Errorf("ipv6 addr should warpped []: %w", ErrInvalidHostAlias)
		}

		ha.IP = net.ParseIP(x[1:end])

		x = x[end:]

		end = strings.IndexByte(x, ':')
		if end == -1 {
			return nil, fmt.Errorf(" ip should end with ':': %w", ErrInvalidHostAlias)
		}
		x = x[end+1:]
	} else {
		end := strings.IndexByte(x, ':')
		if end == -1 {
			return nil, fmt.Errorf(" ip should end with ':': %w", ErrInvalidHostAlias)
		}
		ha.IP = net.ParseIP(x[0:end])
		x = x[end+1:]
	}

	if x == "" {
		return nil, fmt.Errorf("invalid host alias")
	}

	ha.Hostnames = strings.Split(x, ",")

	return ha, nil
}

type HostAlias struct {
	IP        net.IP
	Hostnames []string
}

func (x HostAlias) IsZero() bool {
	return len(x.Hostnames) == 0 || len(x.IP) == 0
}

func (x *HostAlias) UnmarshalText(raw []byte) error {
	ha, err := ParseHostAlias(string(raw))
	if err != nil {
		return err
	}
	*x = *ha
	return nil
}

func (x HostAlias) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

func (x HostAlias) String() string {
	s := strings.Builder{}

	if ip := x.IP.To4(); ip != nil {
		s.WriteString(ip.String())
	} else {
		s.WriteString("[")
		s.WriteString(x.IP.String())
		s.WriteString("]")
	}

	s.WriteString(":")

	for i, hostname := range x.Hostnames {
		if i > 0 {
			s.WriteString(",")
		}
		s.WriteString(hostname)
	}

	return s.String()
}

type Hosts map[string]map[string]struct{}

func (hosts Hosts) WrapDialContext(dialContext func(ctx context.Context, network string, address string) (net.Conn, error)) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		if len(hosts) == 0 {
			return dialContext(ctx, network, addr)
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
			port = "80"
		}

		if ips, ok := hosts[host]; ok && len(ips) > 0 {
			resolved := net.JoinHostPort(hosts.selectIP(maps.Keys(ips), len(ips)), port)
			return dialContext(ctx, network, resolved)
		}

		return dialContext(ctx, network, addr)
	}
}

func (hosts Hosts) selectIP(ips iter.Seq[string], n int) (ip string) {
	i := 0
	idx := 0
	if n > 1 {
		idx = randv2.IntN(n) - 1
	}

	for x := range ips {
		if i == idx {
			ip = x
			break
		}
		i++
	}

	return ip
}

func (hosts Hosts) AddHostAlias(alias HostAlias) {
	if alias.IsZero() {
		return
	}

	for _, hostname := range alias.Hostnames {
		if hosts[hostname] == nil {
			hosts[hostname] = make(map[string]struct{})
		}
		hosts[hostname][alias.IP.String()] = struct{}{}
	}
}
