package event

import (
	"fmt"
	"net"
	"net/http"
)

type MinimalHttpServer struct {
	Mux    *http.ServeMux
	Server *http.Server
	URL    string
}

func (m *MinimalHttpServer) close() error {
	return m.Server.Close()
}

func (m *MinimalHttpServer) handle(path string, handler http.Handler) *MinimalHttpServer {
	m.Mux.Handle(path, handler)
	return m
}

func NewMinimalHttpServer() *MinimalHttpServer {
	addr := fmt.Sprintf("localhost:0")
	mux := http.NewServeMux()
	server := &http.Server{Addr: addr, Handler: mux}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go server.Serve(listener)
	port := listener.Addr().(*net.TCPAddr).Port
	return &MinimalHttpServer{mux, server, fmt.Sprintf("http://localhost:%d", port)}
}
