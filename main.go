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

	"hl.place/resolver/backend"
	hlnames "hl.place/resolver/hlnames"
	hlutil "hl.place/resolver/util"

	"github.com/labstack/echo/v4"
)

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
	e := backend.SetupEcho()

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
	nameHash := hlutil.NameHash(name)
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

	hlutil.PipeURLToResponse(c.Request().Method, url, c.Response())

	latency := time.Since(c.Get("startTime").(time.Time))
	log.Printf("%s (%s) %s --> %s", c.Request().Method, latency, name, url)

	return nil
}

func main() {
	log.Println("Starting server...")
	hlutil.InitLogger()
	hlutil.ResetCache()

	go runHTTPSServer()
	go runHTTPServer()

	log.Println("Server started on https://0.0.0.0:443")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Clean up
	log.Println("Shutting down servers...")
	hlutil.CloseRedisClient()
}
