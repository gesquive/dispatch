package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	log "github.com/sirupsen/logrus"
)

// Server is the dispatch server
type Server struct {
	dispatch *Dispatch
}

// NewServer creates a new dispatch server
func NewServer(dispatch *Dispatch, limitMax float64, limitTTL time.Duration) *Server {
	s := new(Server)
	s.dispatch = dispatch

	// setup a rate limiter if needed
	if limitMax != math.MaxFloat64 {
		log.Infof("setting webserver rate-limit to %d/%s", limitMax, limitTTL)
		lmt := tollbooth.NewLimiter(limitMax, &limiter.ExpirableOptions{DefaultExpirationTTL: limitTTL})
		lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
		lmt.SetMethods([]string{"POST"})

		// setup endpoints
		http.Handle("/send", LimitFuncHandler(lmt, send))
	} else {

		http.HandleFunc("/send", send)
	}
	http.HandleFunc("/", defaultAction)

	return s
}

// Run the server
func (s Server) Run(address string) {
	log.Infof("starting webserver on %s", address)
	log.Fatal(http.ListenAndServe(address, WriteLogHandler(http.DefaultServeMux)))
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

// WriteLogHandler returns a server log handler
func WriteLogHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := statusWriter{w, 0, 0}

		// calculate the latency
		t := time.Now()
		handler.ServeHTTP(&writer, r)
		latency := time.Since(t)

		clientIP := getClientIP(r)
		statusCode := writer.statusCode
		path := r.URL.Path
		method := r.Method
		log.Printf("%s - %s %s %d %v", clientIP, method, path, statusCode, latency)
	})
}

// LimitHandler is a middleware that performs rate-limiting given http.Handler struct.
func LimitHandler(lmt *limiter.Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		httpError := tollbooth.LimitByRequest(lmt, w, r)
		if httpError != nil {
			respondError(w, r, httpError.StatusCode, httpError.Message)
			return
		}

		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}

// LimitFuncHandler is a middleware that performs rate-limiting given request handler function.
func LimitFuncHandler(lmt *limiter.Limiter, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return LimitHandler(lmt, http.HandlerFunc(nextFunc))
}

func send(w http.ResponseWriter, r *http.Request) {
	recvTime := time.Now()
	if r.Method != "POST" {
		respondError(w, r, 404, "page not found")
		return
	}

	if r.Body == nil {
		respondError(w, r, 400, "request body missing")
		return
	}

	requestData := DispatchRequest{}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, r, 400, "message format: %v", err)
		return
	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		respondError(w, r, 400, "message format: %v", err)
		return
	}

	formattedMsg := DispatchRequest{}
	for key, val := range requestData {
		formattedMsg[strings.ToLower(key)] = val
	}
	requestData = formattedMsg
	requestData["timestamp"] = recvTime.Format("Jan 02, 2006 15:04:05 UTC")

	headerData := getHeaderValues(r.Header)

	requestData = DispatchRequest(mergeRequests(headerData, requestData))

	if _, ok := requestData["auth-token"]; !ok {
		respondError(w, r, 400, "'auth-token' missing")
		return
	}

	email, err := FormatEmail(requestData["email"])
	if err != nil {
		respondError(w, r, 400, "email address is not valid")
		return
	}
	requestData["email"] = email

	err = dispatch.Send(requestData)
	if err != nil {
		respondError(w, r, 400, "%v", err)
		return
	}

	respondSuccess(w, r)
}

func getHeaderValues(h http.Header) DispatchRequest {
	headers := DispatchRequest{}
	for header, values := range h {
		if strings.Contains(header, "X-Dispatch-") {
			// we need to go from "X-Dispatch-Auth-Token" to "auth-token"
			c := strings.Replace(header, "X-Dispatch-", "", 1)
			n := strings.ToLower(c)
			for _, value := range values {
				headers[n] = value
			}
		}
	}
	return headers
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

func splitIPList(ipList string) []string {
	ips := strings.Split(ipList, ", ")
	var list []string
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if len(ip) > 0 {
			list = append(list, ip)
		}
	}
	return list
}

func getClientIP(r *http.Request) string {
	// first, figure out the correct IP to use
	clientHostPort := r.RemoteAddr
	proxyList := splitIPList(r.Header.Get("X-Forwarded-For"))
	if len(proxyList) > 0 {
		clientHostPort = proxyList[0]
	}

	// clean it up
	clientIP, _, err := net.SplitHostPort(clientHostPort)
	if err != nil {
		clientIP = clientHostPort
	}

	return clientIP
}
