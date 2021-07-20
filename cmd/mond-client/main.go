package main

import (
	"fmt"
	mond "mond-api"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	reportUrl := "http://localhost:5000/health/Test"
	websites := []string{
		"http://localhost:5001/",
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	//go func() {
	for {
		select {
		case <-ticker.C:
			// do check
			results := mond.CheckWebsites(mond.CheckWebsite, websites)
			for k, v := range results {
				status, err := mond.ReportHealthCheck(mond.Report, reportUrl, v)
				if err != nil {
					fmt.Printf("problem reporting health: %v", err)
				}
				if status != http.StatusAccepted {
					fmt.Printf("problem reporting health, got status=%d want 202 \n", status)
				} else {
					fmt.Printf("reported %s=%v \n", k, v)
				}
			}
		case <-quit:
		case <-c:
			ticker.Stop()
			fmt.Println("Quit")
			return
		}
	}
	//}()
}
