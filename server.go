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
	http.HandleFunc("/", defaultAction)

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
		respondError(w, r, 404, "page not found")
		return
	}

	if r.Body == nil {
		respondError(w, r, 400, "request body missing")
		return
	}
	var msg DispatchRequest
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		respondError(w, r, 400, "message format: %v", err)
		return
	}

	if len(msg.AuthToken) == 0 {
		respondError(w, r, 400, "field 'auth-token' missing")
		return
	}

	err = dispatch.Send(msg)
	if err != nil {
		respondError(w, r, 400, "%v", err)
		return
	}

	respondSuccess(w, r)
}

func defaultAction(w http.ResponseWriter, r *http.Request) {
	respondError(w, r, 404, "page not found")
}

func respondError(w http.ResponseWriter, r *http.Request, code int, message string, a ...interface{}) {
	var msg string
	if r.Header.Get("Content-Type") == "application/json" {
		m := fmt.Sprintf(message, a...)
		msg = fmt.Sprintf("{\"status\": \"error\", \"message\": \"%s\"}", m)
		w.Header().Add("Content-Type", "application/json")
	} else { // default is text response
		m := fmt.Sprintf(message, a...)
		msg = fmt.Sprintf("%d %s", code, m)
		w.Header().Add("Content-Type", "text/plain")
	}

	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondSuccess(w http.ResponseWriter, r *http.Request) {
	var msg string
	if r.Header.Get("Content-Type") == "application/json" {
		msg = "{\"status\": \"success\"}"
		w.Header().Add("Content-Type", "application/json")
	} else { // default is text response
		msg = "200 success"
		w.Header().Add("Content-Type", "text/plain")
	}

	w.WriteHeader(200)
	w.Write([]byte(msg))
}
