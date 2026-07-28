package network

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Manager struct {
	mu        sync.Mutex
	port      int
	logger    *slog.Logger
	listeners []net.Listener
	urls      []string
}

func NewManager(port int, logger *slog.Logger) *Manager {
	return &Manager{port: port, logger: logger}
}

func (m *Manager) Enable(handler http.Handler) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.urls) > 0 {
		return append([]string(nil), m.urls...), nil
	}
	addresses, err := privateAddresses()
	if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, m.port))
		if err != nil {
			m.logger.Warn("could not bind LAN address", "address", address, "error", err)
			continue
		}
		m.listeners = append(m.listeners, listener)
		m.urls = append(m.urls, fmt.Sprintf("http://%s:%d", address, m.port))
		server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
		go func(l net.Listener) {
			if serveErr := server.Serve(l); serveErr != nil &&
				!errors.Is(serveErr, http.ErrServerClosed) &&
				!errors.Is(serveErr, net.ErrClosed) {
				m.logger.Error("LAN server stopped", "error", serveErr)
			}
		}(listener)
	}
	if len(m.urls) == 0 {
		return nil, errors.New("no LAN interface could be opened")
	}
	return append([]string(nil), m.urls...), nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, listener := range m.listeners {
		_ = listener.Close()
	}
	m.listeners = nil
	m.urls = nil
}

func privateAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range addresses {
			var ip net.IP
			switch value := raw.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || ip.To4() == nil || !ip.IsPrivate() {
				continue
			}
			result = append(result, ip.String())
		}
	}
	preferred := preferredLocalIPv4()
	sort.SliceStable(result, func(i, j int) bool {
		if result[i] == preferred && result[j] != preferred {
			return true
		}
		if result[j] == preferred && result[i] != preferred {
			return false
		}
		return result[i] < result[j]
	})
	return result, nil
}

func preferredLocalIPv4() string {
	connection, err := net.Dial("udp4", "192.0.2.1:80")
	if err != nil {
		return ""
	}
	defer connection.Close()
	address, ok := connection.LocalAddr().(*net.UDPAddr)
	if !ok || address.IP.To4() == nil {
		return ""
	}
	return address.IP.String()
}
