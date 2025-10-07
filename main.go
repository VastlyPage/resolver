package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	hlnames "vastly.page/hl.baby/hlnames"
	hlbabyutil "vastly.page/hl.baby/util"

	"github.com/fatih/color"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func initLogger() {
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

// HTTPS CONNECT tunneling handler
func handleConnect(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		clientConn.Close()
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go func() { io.Copy(targetConn, clientConn); targetConn.Close() }()
	go func() { io.Copy(clientConn, targetConn); clientConn.Close() }()
}

func runHTTPSServer() {
	e := echo.New()
	e.Use(middleware.Gzip(), middleware.Recover())
	e.HideBanner = true
	e.Logger.SetOutput(io.Discard)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("startTime", time.Now())
			return next(c)
		}
	})

	e.Any("/*", handleRequest)

	// Use standard net/http mux for CONNECT
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			handleConnect(w, r)
			return
		}
		e.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:     ":443",
		Handler:  mux,
		ErrorLog: log.New(io.Discard, "", 0), // Suppress TLS handshake errors
	}
	log.Fatal(server.ListenAndServeTLS("fullchain.pem", "privkey.pem"))
}

func runHTTPServer() {
	httpRedirect := echo.New()
	httpRedirect.GET("/*", func(c echo.Context) error {
		target := "https://" + c.Request().Host + c.Request().URL.Path
		log.Printf("Redirecting HTTP to HTTPS: %s", target)
		return c.Redirect(http.StatusMovedPermanently, target)
	})
	httpRedirect.Logger.Fatal(httpRedirect.Start(":80"))
}

func handleRequest(c echo.Context) error {
	hostParts := strings.Split(c.Request().Host, ".")
	if len(hostParts) < 2 || (!strings.HasSuffix(c.Request().Host, "hl.place") && !strings.HasSuffix(c.Request().Host, "hl.baby") && !strings.HasSuffix(c.Request().Host, ".hl")) {
		return c.Redirect(http.StatusFound, "https://vastly.page")
	}

	// TODO: Subdomain subdomain subdomain. e.g. subsub.sub.example.hl.place
	subdomain := hostParts[0]
	name := subdomain + ".hl"
	nameHash := hlbabyutil.NameHash(name)
	kv := hlnames.QueryKV(nameHash)

	if kv == nil {
		log.Printf("No data found for %s", name)
		return c.String(http.StatusNotFound, "No data found")
	}

	_, url, isRedirect, err := hlnames.ResolveHostAndURL(kv, c.Request().URL.Path)
	if err != nil {
		log.Printf("%s: %v", name, err)
		return c.String(http.StatusServiceUnavailable, err.Error())
	}

	if isRedirect {
		return c.Redirect(http.StatusFound, url)
	}

	hlbabyutil.PipeURLToResponse(c.Request().Method, url, c.Response())
	latency := time.Since(c.Get("startTime").(time.Time))
	log.Printf("%s (%s) %s --> %s", c.Request().Method, latency, name, url)
	return nil
}

func main() {
	initLogger()
	hlbabyutil.ResetCache()

	go runHTTPSServer()
	go runHTTPServer()

	log.Println("Server started on https://0.0.0.0:443")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down servers...")

	// Close Redis client on shutdown
	hlbabyutil.CloseRedisClient()
}
