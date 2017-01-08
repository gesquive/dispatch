package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/didip/tollbooth"
)

// Server is the dispatch server
type Server struct {
	dispatch *Dispatch
}

// NewServer creates a new dispatch server
func NewServer(dispatch *Dispatch, limitMax int64, limitTTL time.Duration) *Server {
	s := new(Server)
	s.dispatch = dispatch

	// setup a rate limiter if needed
	if limitMax != math.MaxInt64 {
		log.Infof("setting webserver rate-limit to %d/%s", limitMax, limitTTL)
		limiter := tollbooth.NewLimiter(limitMax, limitTTL)
		limiter.IPLookups = []string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"}
		limiter.Methods = []string{"POST"}

		// setup endpoints
		http.Handle("/send", tollbooth.LimitFuncHandler(limiter, send))
	} else {

		http.HandleFunc("/send", send)
	}

	return s
}

// Run the server
func (s Server) Run(address string) {
	log.Infof("starting webserver on %s", address)
	log.Fatal(http.ListenAndServe(address, WriteLog(http.DefaultServeMux)))
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func (w *statusWriter) WriteHeader(status int) {
	w.statusCode = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

// WriteLog returns a server log handler
func WriteLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := statusWriter{w, 0, 0}

		// calculate the latency
		t := time.Now()
		handler.ServeHTTP(&writer, r)
		latency := time.Since(t)

		clientIP := r.RemoteAddr
		statusCode := writer.statusCode
		path := r.URL.Path
		method := r.Method
		log.Printf("%s - %s %s %d %v", clientIP, method, path, statusCode, latency)
	})
}

func send(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "404 page not found", 404)
		return
	}

	if r.Body == nil {
		http.Error(w, makeResponse("error", "please send a request body"), 400)
		return
	}
	var msg DispatchRequest
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, makeResponse("error", "message format: %v", err), 400)
		return
	}

	if len(msg.AuthToken) == 0 {
		http.Error(w, makeResponse("error", "field 'auth-token' missing or incomplete"), 400)
		return
	}

	err = dispatch.Send(msg)
	if err != nil {
		http.Error(w, makeResponse("error", "%v", err), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(makeResponse("success", "")))
}

func makeResponse(status string, message string, a ...interface{}) string {
	if len(message) == 0 {
		return fmt.Sprintf("{\"status\": \"%s\"}", status)
	}
	msg := fmt.Sprintf(message, a...)
	return fmt.Sprintf("{\"status\": \"%s\", \"message\": \"%s\"}", status, msg)
}
