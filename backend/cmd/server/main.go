package main

import (
	"context"
	"net"
	"net/http"
)

type Config struct {
	Port string
	Host string
}

func RunServer(ctx context.Context, cfg *Config, ready chan<- struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	close(ready)
	return srv.ListenAndServe()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	RunServer(ctx, &Config{Port: "8080", Host: "localhost"}, ready)
}
