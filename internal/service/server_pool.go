package service

import (
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/tty-monkey/load-balancer/internal/models"
)

type ServerPool interface {
	AddServer(server *models.Server)
	MarkServerStatus(url *url.URL, alive bool)
	GetNextServer() *models.Server
	IsServerAlive(url *url.URL) bool
	HealthCheck()
}

type DefaultServerPool struct {
	servers []*models.Server
	current uint64
}

func NewServerPool() ServerPool {
	return &DefaultServerPool{
		servers: make([]*models.Server, 0),
		current: 0,
	}
}

func (s *DefaultServerPool) AddServer(server *models.Server) {
	s.servers = append(s.servers, server)
}

func (s *DefaultServerPool) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.servers)))
}

func (s *DefaultServerPool) MarkServerStatus(url *url.URL, alive bool) {
	for _, server := range s.servers {
		if server.URL.String() == url.String() {
			server.SetAlive(alive)
			break
		}
	}
}

func (s *DefaultServerPool) IsServerAlive(url *url.URL) bool {
	for _, server := range s.servers {
		if server.URL.String() == url.String() {
			return server.IsAlive()
		}
	}
	return false
}

func (s *DefaultServerPool) GetNextServer() *models.Server {
	next := s.nextIndex()
	l := len(s.servers) + next
	for i := next; i < l; i++ {
		idx := i % len(s.servers)
		if s.servers[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.servers[idx]
		}
	}
	return nil
}

func (s *DefaultServerPool) HealthCheck() {
	for _, server := range s.servers {
		status := "up"
		alive := isServerAlive(server.URL)
		server.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", server.URL, status)
	}
}

func isServerAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}
