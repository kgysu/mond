package main

import (
	"fmt"
	mond "mond-api"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Too few arguments")
		return
	}

	reportUrl := args[0]
	_, err := url.ParseRequestURI(reportUrl)
	if err != nil {
		fmt.Printf("Invalid URL: %s", err.Error())
		return
	}
	websites := args[1:]
	fmt.Printf("Reporting health to %s \n from %v \n", reportUrl, websites)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
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
}
