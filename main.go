package main

import (
	"flag"
	"log"
	"os"
	"pi_temperature_monitor/vcgencmd"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	samplingInterval = 10 * time.Second
	rotationInterval = 12 * time.Hour
)

var (
	signal = flag.String("s", "", `Send signal to the daemon:
	quit — graceful shutdown
	stop — fast shutdown
	reload — reloading the configuration file`)

	ticker      *time.Ticker
	duration    time.Duration
	pidFileName string

	logger  *lumberjack.Logger
	logSync sync.Mutex
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, quitHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, stopHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	setupLogging()

	cntxt := &daemon.Context{
		PidFileName: "pi-monitoring.pid",
		PidFilePerm: 0644,
		// LogFileName: "pi-monitoring.log",
		// LogFilePerm: 0640,
		WorkDir: "./",
		Umask:   027,
		Args:    []string{"[pi-monitoring ]"},
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
	for range t.C {
		logSync.Lock()
		vcgencmd.NewCmd().
			SetSubcmd(vcgencmd.MEASURE_TEMP).
			SetSubcmd(vcgencmd.GET_THROTTLED).
			Run()
		logSync.Unlock()
	}
}

func readConfig() {
	const configFile = "pi-monitoring.conf"
	if _, err := os.Stat(configFile); err != nil {
		os.WriteFile("pi-monitoring.conf", []byte("ticker_interval=10s\n"), 0644)
		duration = samplingInterval
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
	logger = &lumberjack.Logger{
		Filename:   "pi-monitoring.log",
		MaxSize:    4, // 4MB
		MaxAge:     28,
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}
	log.SetOutput(logger)
}

func quitHandler(sig os.Signal) error {
	logSync.Lock()
	defer logSync.Unlock()
	// os.RemoveAll(pidFileName)
	ticker.Stop()
	return nil
}

func stopHandler(sig os.Signal) error {
	return quitHandler(sig)
}

func reloadHandler(sig os.Signal) error {
	logSync.Lock()
	defer logSync.Unlock()
	log.Println("rotating")
	logger.Rotate()
	readConfig()
	ticker.Reset(duration)
	return nil
}
