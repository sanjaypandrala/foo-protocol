package main

import (
	"os/signal"
	"os"
	"syscall"
	"flag"
)

func main() {
	lAddr := flag.String("listen", ":8002", "Address to listen")
	fAddr := flag.String("forward", "localhost:8001", "Destination server")
	mAddr := flag.String("metrics", ":8003", "Metrics address")
	flag.Parse()

	proxy := &Proxy{}
	go proxy.Execute(*lAddr, *fAddr)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGUSR1)
	go MetricsInstance.ResolveSignals(sigc)

	MetricsInstance.Execute(*mAddr)
}
