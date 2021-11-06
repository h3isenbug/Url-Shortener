package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/h3isenbug/url-shortener/cmd/url-shortener/di"
)

func main() {
	app, cleanup, err := di.Inject()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialize app: %s\n", err.Error())
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)

	go func() {
		<-c
		cleanup()
	}()
	app.Start()
}
