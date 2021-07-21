package main

import (
	"bufio"
	"fmt"
	mond "mond-api"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const MondStartCmd = "MOND_START_CMD"
const AppName = "test"

func checkEnv() (string, string, error) {
	os.Setenv(MondStartCmd, "ping 127.0.0.1") // TODO remove, used for testing only
	startCmdWithArgs := os.Getenv(MondStartCmd)
	if startCmdWithArgs == "" {
		return "", "", fmt.Errorf("no start command defined, set %s for current env", MondStartCmd)
	}
	fmt.Printf("MOND_START_CMD: %s\n", startCmdWithArgs)
	scSplitted := strings.SplitN(startCmdWithArgs, " ", 2)
	startCmd := scSplitted[0]
	startArgs := ""
	if len(scSplitted) > 1 {
		startArgs = scSplitted[1]
	}
	return startCmd, startArgs, nil
}

func checkArgs() (string, []string, error) {
	args := os.Args[1:]
	if len(args) < 2 {
		return "", nil, fmt.Errorf("Too few arguments \n")
	}

	reportUrl := args[0]
	_, err := url.ParseRequestURI(reportUrl)
	if err != nil {
		return "", nil, fmt.Errorf("Invalid URL: %s \n", err.Error())
	}
	websites := args[1:]
	return reportUrl, websites, nil
}

func main() {
	startCmd, startArgs, err := checkEnv()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Start Command: %s (%v)\n ", startCmd, startArgs)

	reportUrl, websites, err := checkArgs()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Reporting to %s \n from %v \n", reportUrl, websites)

	// Start watching Command Stdout
	go startAndWatchStdout(reportUrl, startCmd, startArgs)
	reportHealthUrl := reportUrl + mond.ApiHealthPath + AppName

	// Start reporting health
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
				status, err := mond.ReportHealthCheck(mond.Report, reportHealthUrl, v)
				if err != nil {
					fmt.Printf("problem reporting health: %v", err)
				}
				if status != http.StatusAccepted {
					fmt.Printf("problem reporting health, got status=%d want 202 \n", status)
				} else {
					// successfully reported TODO check if needed
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

func startAndWatchStdout(reportUrl, command, args string) {
	reportLogsUrl := reportUrl + mond.ApiAccessLogsPath + AppName
	cmd := exec.Command(command, args)
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(out)
	stop := make(chan bool)
	go readStuff(reportLogsUrl, scanner, stop)
	<-stop

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}

func readStuff(reportUrl string, scanner *bufio.Scanner, stop chan bool) {
	for scanner.Scan() {
		fmt.Println("Performed Scan")
		fmt.Println(scanner.Text())
		err := mond.ReportRawLog(reportUrl, scanner.Text())
		if err != nil {
			fmt.Println(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	stop <- true
}


