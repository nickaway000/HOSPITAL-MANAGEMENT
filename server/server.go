
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go startAppointmentServer()
	go startPharmacyServer()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servers...")
}
