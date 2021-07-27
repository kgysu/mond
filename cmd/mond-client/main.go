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

const MondStartCmdEnv = "MOND_START_CMD"
const MondAppNameEnv = "MOND_APP_NAME"

func checkEnv() (string, string, error) {
	//os.Setenv(MondStartCmd, "ping 127.0.0.1") // TODO remove, used for testing only
	startCmdWithArgs := os.Getenv(MondStartCmdEnv)
	if startCmdWithArgs == "" {
		return "", "", fmt.Errorf("no start command defined, set %s for current env", MondStartCmdEnv)
	}
	scSplitted := strings.SplitN(startCmdWithArgs, " ", 2)
	startCmd := scSplitted[0]
	startArgs := ""
	if len(scSplitted) > 1 {
		startArgs = scSplitted[1]
	}
	return startCmd, startArgs, nil
}

func getAppNameFromEnv() string {
	appName := os.Getenv(MondAppNameEnv)
	if appName == "" {
		appName = "default"
		fmt.Println("WARN: using default app name!")
	}
	return appName
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
	appName := getAppNameFromEnv()

	startCmd, startArgs, err := checkEnv()
	if err != nil {
		fmt.Printf("ERROR: %v \n", err)
		return
	}

	reportUrl, websites, err := checkArgs()
	if err != nil {
		fmt.Printf("ERROR: %v \n", err)
		return
	}
	fmt.Printf("Reporting to %s \n from %v \n", reportUrl, websites)

	// check ReportUrl
	err = mond.ReportRawLog(reportUrl, "Start Reporting")
	if err != nil {
		fmt.Printf("ERROR: Reporting: %v \n", err)
	}

	// Start reporting health
	go startReportingHealth(appName, reportUrl, websites)

	// Start command and watching Stdout
	err = startCmdAndWatchStdout(appName, reportUrl, startCmd, startArgs)
	if err != nil {
		fmt.Printf("ERROR: %v \n", err)
	}
}

func startCmdAndWatchStdout(appName, reportUrl, command, args string) error {
	reportLogsUrl := reportUrl + mond.ApiAccessLogsPath + appName
	var argsArr []string
	if strings.Contains(args, "'") {
		argsArr = strings.SplitN(args, " ", 2)
		argsArr[1] = strings.ReplaceAll(argsArr[1], "'", "")
	} else {
		argsArr = strings.Split(args, " ")
	}
	fmt.Printf("Start Command: %s\n ", command)
	for i, c := range argsArr {
		fmt.Printf("- Arg %d: %s\n ", i, c)
	}
	cmd := exec.Command(command, argsArr...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errOut, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(out)
	errScanner := bufio.NewScanner(errOut)
	stop := make(chan bool)
	go readStuff(reportLogsUrl, scanner, stop)
	go readStuff(reportLogsUrl, errScanner, stop)
	<-stop

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func startReportingHealth(appName, reportUrl string, websites []string) {
	reportHealthUrl := reportUrl + mond.ApiHealthPath + appName
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			// do check
			results := mond.CheckWebsites(mond.CheckWebsite, websites)
			for _, v := range results {
				status, err := mond.ReportHealthCheck(mond.Report, reportHealthUrl, v)
				if err != nil {
					fmt.Printf("problem reporting health: %v\n", err)
				}
				if status != http.StatusAccepted {
					fmt.Printf("got status=%d want 202 \n", status)
				} else {
					// successfully reported TODO check if needed
					//fmt.Printf("reported %s=%v \n", k, v)
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

func readStuff(reportUrl string, scanner *bufio.Scanner, stop chan bool) {
	for scanner.Scan() {
		//fmt.Println("Performed Scan")
		text := scanner.Text()
		fmt.Println(text)
		mond.ReportRawLog(reportUrl, text)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading standard input:", err)
	}
	stop <- true
}
