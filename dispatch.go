package main

import (
	"gopkg.in/gin-gonic/gin.v1"

	log "github.com/Sirupsen/logrus"
)

type dispatch struct {
	AuthToken string
	From      string
	To        []string
}

// DispatchMessage is the expected message submission
type DispatchMessage struct {
	AuthToken string `json:"auth-token"`
	Message   string `json:"message"`
	Subject   string `json:"subject"`
}

func send(c *gin.Context) {
	var msg DispatchMessage
	c.BindJSON(&msg)

	if len(msg.AuthToken) == 0 {
		c.JSON(400, gin.H{"status": "error: field 'auth-token' missing or incomplete"})
		return
	}

	target, found := dispatchMap[msg.AuthToken]
	if !found {
		c.JSON(401, gin.H{"status": "error: auth-token not recognized"})
		return
	}

	var email Message
	email.FromAddress = target.From
	email.ToAddressList = target.To
	email.TextMessage = msg.Message
	email.Subject = msg.Subject

	log.Debugf("sending message: %+v", email)

	sendMessage(email, smtpSettings)
	c.JSON(200, gin.H{"status": "success"})
}
