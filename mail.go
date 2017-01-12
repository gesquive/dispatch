package main

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"

	log "github.com/Sirupsen/logrus"
)

// Message defines a message to send
type Message struct {
	FromAddress   string
	ToAddressList []string
	Subject       string
	TextMessage   string
	HTMLMessage   string
}

// SMTPSettings defines an SMTP server settings
type SMTPSettings struct {
	Host     string
	Port     int
	UserName string
	Password string
}

func sendMessage(message Message, smtp SMTPSettings) (success bool) {
	success = false
	msg := gomail.NewMessage()
	log.Debugf("Date: %s", time.Now().Format(time.RFC1123Z))

	if len(message.FromAddress) == 0 {
		message.FromAddress = getDefaultEmailAddress()
	}
	log.Debugf("From: %s", message.FromAddress)
	msg.SetHeader("From", message.FromAddress)

	toAddresses, err := formatEmailList(message.ToAddressList)
	if err != nil {
		log.Warn("%v", err)
		log.Error("Will not send email")
		return
	} else if len(toAddresses) > 0 {
		log.Debugf("To: %s", strings.Join(toAddresses, ", "))
		msg.SetHeader("To", toAddresses...)
	}

	log.Debugf("Subject: %s", message.Subject)
	msg.SetHeader("Subject", message.Subject)

	haveText := len(message.TextMessage) > 0
	haveHTML := len(message.HTMLMessage) > 0
	if haveText && haveHTML {
		log.Debugf("Content-Type: text/plain")
		log.Debugf("Message: %s", message.TextMessage)
		msg.SetBody("text/plain", message.TextMessage)
		log.Debugf("Content-Type: text/html")
		log.Debugf("Message: %s", message.HTMLMessage)
		msg.AddAlternative("text/html", message.HTMLMessage)
	} else if haveText {
		log.Debugf("Content-Type: text/plain")
		log.Debugf("Message: %s", message.TextMessage)
		msg.SetBody("text/plain", message.TextMessage)
	} else if haveHTML {
		log.Debugf("Content-Type: text/html")
		log.Debugf("Message: %s", message.HTMLMessage)
		msg.SetBody("text/html", message.HTMLMessage)
	} else {
		log.Warn("There is no message to send")
		return
	}

	//TODO: Add an option for tls/ssl connections
	var dialer *gomail.Dialer
	if len(smtp.UserName) > 0 || len(smtp.Password) > 0 {
		log.Debugf("Connecting too %s:*****@%s:%d", smtp.UserName, smtp.Host, smtp.Port)
		dialer = gomail.NewDialer(smtp.Host, smtp.Port, smtp.UserName, smtp.Password)
	} else {
		log.Debugf("Connecting too %s:%d", smtp.Host, smtp.Port)
		dialer = &gomail.Dialer{Host: smtp.Host, Port: smtp.Port}
	}
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(msg); err != nil {
		log.Error("An error occurred when sending email")
		log.Error(err)
		return false
	}
	success = true
	return success
}

func formatEmailList(list []string) ([]string, error) {
	var formattedList []string
	for _, r := range list {
		formattedAddress, err := formatEmail(r)
		if err != nil {
			return []string{},
				fmt.Errorf("Could not parse address '%s': %v", r, err)
		}
		formattedList = append(formattedList, formattedAddress)
	}
	return formattedList, nil
}

func formatEmail(address string) (string, error) {
	email, err := mail.ParseAddress(address)
	if err != nil {
		return "", err
	}

	var fAddress string
	if len(email.Name) > 0 {
		fAddress = fmt.Sprintf("\"%s\" <%s>", email.Name, email.Address)
	} else {
		fAddress = email.Address
	}
	return fAddress, nil
}

func getDefaultEmailAddress() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("dispatch@%s", hostname)
}
