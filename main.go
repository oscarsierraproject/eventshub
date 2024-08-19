package main

import (
	v1rest "eventshub/service/v1/rest"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var wg sync.WaitGroup

	restServer := v1rest.HTTPRestServer{}

	// We want a server to gracefully shutdown after receiving
	// a SIGTERM, or a SIGINT (Ctrl+C) signal.
	sigs := make(chan os.Signal, 1)

	restServer.Configure(sigs)
	restServer.StartTLS()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()

		sig := <-sigs
		log.Printf("Received %s signal, terminating.\n", sig)

		err := restServer.Stop()
		if err != nil {
			log.Fatalln(err)
		}

	}()
	wg.Wait() // The program will wait here until Ctrl+C is pressed.
}
