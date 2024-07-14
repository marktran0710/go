package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	mux   sync.RWMutex
}

type LoadBalancer struct {
	servers []*Backend
	current uint64
}

func NewLoadBalancer(servers []string) (*LoadBalancer, error) {
	backendServers := make([]*Backend, len(servers))
	fmt.Println(len(servers))
	for i, server := range servers {
		url, err := url.Parse(server)
		if err != nil {
			return nil, err
		}
		fmt.Println(server, i)
		backendServers[i] = &Backend{
			URL: url,
		}
	}
	return &LoadBalancer{servers: backendServers}, nil
}

func (lb *LoadBalancer) getNextServer() *Backend {
	next := atomic.AddUint64(&lb.current, 1)
	return lb.servers[(next-1)%uint64(len(lb.servers))]
}

func (lb *LoadBalancer) serveReverseProxy(target *url.URL, res http.ResponseWriter, req *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(target)

	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.Host = target.Host

	proxy.ServeHTTP(res, req)
}

func (lb *LoadBalancer) handleRequest(res http.ResponseWriter, req *http.Request) {
	target := lb.getNextServer()
	fmt.Printf("Forwarding request to %s\n", target.URL)
	lb.serveReverseProxy(target.URL, res, req)
}

func (be *Backend) SetAlive(alive bool) {
	be.mux.Lock()
	be.Alive = alive
	be.mux.Unlock()
}

// HealthCheck pings the backends and update the status
func (lb *LoadBalancer) HealthCheck() {
	for _, b := range lb.servers {
		status := "up"
		alive := isBackendAlive(b)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

var lb LoadBalancer

// healthCheck runs a routine for check status of the backends every 2 mins
func healthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			lb.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

// isAlive checks whether a backend is Alive by establishing a TCP connection
func isBackendAlive(be *Backend) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", be.URL.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

func main() {
	servers := []string{
		"https://grafana.com:443",
		"http://info.cern.ch:80",
	}

	lb, err := NewLoadBalancer(servers)
	if err != nil {
		log.Fatal("Error creating load balancer: ", err)
	}

	http.HandleFunc("/", lb.handleRequest)

	go healthCheck()

	fmt.Println("Load Balancer started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
