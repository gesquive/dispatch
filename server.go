package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/gin-gonic/gin.v1"
	"time"
)

// Server is the dispatch server
type Server struct {
	dispatch *Dispatch
	router   *gin.Engine
}

// NewServer creates a new dispatch server
func NewServer(dispatch *Dispatch) *Server {
	s := new(Server)
	s.dispatch = dispatch

	router := gin.New()
	s.router = router
	s.router.Use(webLogger)

	s.router.POST("/send", send)

	return s
}

// Run the server
func (s Server) Run(address string) {
	log.Infof("Starting webserver on %s", address)
	s.router.Run(address)
}

func webLogger(c *gin.Context) {
	// calculate the latency
	t := time.Now()
	c.Next()
	latency := time.Since(t)

	clientIP := c.ClientIP()
	statusCode := c.Writer.Status()
	path := c.Request.URL.Path
	method := c.Request.Method

	log.Printf("%s - %s %s %d %v", clientIP, method, path, statusCode, latency)
}

func send(c *gin.Context) {
	var msg DispatchRequest
	c.BindJSON(&msg)

	if len(msg.AuthToken) == 0 {
		c.JSON(400, gin.H{"status": "error",
			"message": "field 'auth-token' missing or incomplete"})
		return
	}

	err := dispatch.Send(msg)
	if err != nil {
		c.JSON(401, gin.H{"status": "error", "message": err})
	}

	c.JSON(200, gin.H{"status": "success"})
}
