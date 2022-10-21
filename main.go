package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddr := os.Getenv("EXPERIA_V10_LISTEN_ADDR")
	timeout, err := time.ParseDuration(os.Getenv("EXPERIA_V10_TIMEOUT"))
	if err != nil {
		log.Fatalf("EXPERIA_V10_TIMEOUT invalid: %s", err)
	}
	ip := net.ParseIP(os.Getenv("EXPERIA_V10_ROUTER_IP"))
	if ip == nil {
		log.Fatal("EXPERIA_V10_ROUTER_IP invalid")
	}
	username := os.Getenv("EXPERIA_V10_ROUTER_USERNAME")
	password := os.Getenv("EXPERIA_V10_ROUTER_PASSWORD")

	collector := newCollector(ip, username, password, timeout)
	if err := prometheus.Register(collector); err != nil {
		log.Fatalf("Failed to register collector: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	log.Printf("Listen on %s...", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
