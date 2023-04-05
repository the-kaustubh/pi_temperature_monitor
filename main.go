package main

import (
	"flag"
	"log"
	"os"
	"pi_temperature_monitor/vcgencmd"
	"strings"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

var (
	signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)

	ticker      *time.Ticker
	duration    time.Duration
	pidFileName string
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, quitHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, stopHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)
	cntxt := &daemon.Context{
		PidFileName: "pi-monitoring.pid",
		PidFilePerm: 0644,
		LogFileName: "pi-monitoring.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[pi-monitoring ]"},
	}

	pidFileName = cntxt.PidFileName

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("Starting daemon - ", os.Getpid())
	// setupLogging()
	readConfig()

	ticker = time.NewTicker(duration)
	go runTicker(ticker)

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("daemon terminated")
}

func runTicker(t *time.Ticker) {
	for i := range t.C {
		_ = i
		vcgencmd.NewCmd().
			SetSubcmd(vcgencmd.MEASURE_TEMP).
			SetSubcmd(vcgencmd.GET_THROTTLED).
			Run()
	}
}

func readConfig() {
	const configFile = "pi-monitoring.conf"
	if _, err := os.Stat(configFile); err != nil {
		os.WriteFile("pi-monitoring.conf", []byte("ticker_interval=5s\n"), 0644)
		duration = 5 * time.Second
		return
	}
	content, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("osReadFileErr: ", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ticker_interval") {
			durationPart := strings.Split(line, "=")[1]
			duration, err = time.ParseDuration(durationPart)
			if err != nil {
				log.Fatal("timeParseDuration", err)
			}
		}
	}

}

func setupLogging() {
	f, err := os.OpenFile("pi-monitoring.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
}

func quitHandler(sig os.Signal) error {
	os.RemoveAll(pidFileName)
	ticker.Stop()
	return nil
}

func stopHandler(sig os.Signal) error {
	os.RemoveAll(pidFileName)
	ticker.Stop()
	return nil
}

func reloadHandler(sig os.Signal) error {
	readConfig()
	ticker.Reset(duration)
	return nil
}
