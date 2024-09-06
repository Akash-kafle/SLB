package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Server struct {
	address string
	proxy   *httputil.ReverseProxy
}

type LoadBalancer struct {
	port           string
	rounRobincount int
	servers        []Server
}
type ServerImpl interface {
	Address() string
	IsAlive() (bool, error)
	Serve(w http.ResponseWriter, r *http.Request)
}

func LoadBalancernew(port string, servers []Server) *LoadBalancer {
	var Alive_servers []Server

	for _, server := range servers {
		check, err := server.IsAlive()
		if check && err == nil {
			Alive_servers = append(Alive_servers, server)
		} else {
			fmt.Printf("Server %s is not available\n Skipping that......\n", server.address)
		}
	}

	return &LoadBalancer{
		port:           port,
		rounRobincount: 0,
		servers:        Alive_servers,
	}
}

func SimpleServer(address string) *Server {
	serverUrl, err := url.Parse(address)
	HandleErr(err)
	return &Server{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func HandleErr(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

func (s *Server) IsAlive() (bool, error) {
	// Set a reasonable timeout value
	timeout := 5 * time.Second

	// Create a new HTTP client
	client := &http.Client{Timeout: timeout}

	// Send a HEAD request to the URL
	req, err := http.NewRequest("HEAD", s.address, nil)
	if err != nil {
		return false, err
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	// If we've reached this point, the URL likely responds with a 200 status code
	return true, nil
	// return true, nil
}

func (s *Server) Address() string {
	return s.address
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	if s == nil {
		return
	}
	s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) GetNextAvailableServer() Server {
	if len(lb.servers) < 1 {
		return Server{}
	}
	server := lb.servers[lb.rounRobincount%len(lb.servers)]
	checker, err := server.IsAlive()
	for !checker && err != nil {
		lb.rounRobincount++
		server = lb.servers[lb.rounRobincount%len(lb.servers)]
	}
	lb.rounRobincount++
	return server
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	targetserver := lb.GetNextAvailableServer()
	fmt.Printf("Forwarding to %q\n", targetserver.address)
	targetserver.Serve(w, r)
}
