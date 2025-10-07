package hlutil

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

type colorWriter struct{}

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	switch {
	case strings.Contains(msg, "[ERROR]"):
		return color.New(color.FgRed).Fprint(os.Stdout, msg)
	case strings.Contains(msg, "[WARN]"):
		return color.New(color.FgYellow).Fprint(os.Stdout, msg)
	case strings.Contains(msg, "[INFO]"):
		return color.New(color.FgCyan).Fprint(os.Stdout, msg)
	default:
		return color.New(color.FgWhite).Fprint(os.Stdout, msg)
	}
}

func InitLogger() {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "[ERROR] Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(io.MultiWriter(logFile, &colorWriter{}))

	// Ensure log file is closed on shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logFile.Close()
	}()
}
