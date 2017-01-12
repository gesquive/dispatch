package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// DispatchMap is a AuthToken to DispatchTarget map
type DispatchMap map[string]DispatchTarget

// DispatchRequest is the expected message submission
type DispatchRequest struct {
	AuthToken string    `json:"auth-token" binding:"required"`
	Message   string    `json:"message"`
	Subject   string    `json:"subject"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Time      time.Time `json:"-"`
}

// Dispatch is the central point for the dispatches
type Dispatch struct {
	dispatchMap     DispatchMap
	smtpSettings    SMTPSettings
	messageTemplate *template.Template
}

// NewDispatch create a new dispatch
func NewDispatch(targetDir string, smtpSettings SMTPSettings) *Dispatch {
	d := new(Dispatch)
	d.dispatchMap = make(DispatchMap)
	d.smtpSettings = smtpSettings
	msg := `
Dispatch message
===========================================================
Timestamp: {{.Time.Format "Jan 02, 2006 15:04:05 UTC"}}
Name:      {{.Name}}
Email:     {{.Email}}
Subject:   {{.Subject}}
-----------------------------------------------------------
{{.Message}}
`

	d.messageTemplate = template.Must(template.New("request").Parse(msg))
	d.LoadTargets(targetDir)
	return d
}

// LoadTargets Loads all of the configs in the target config dir
func (d *Dispatch) LoadTargets(targetDir string) {
	targets, err := getTargetConfigList(targetDir)
	if err != nil {
		log.Errorf("error: could not load targets: %v", err)
		return
	}
	log.Debugf("Found %d targets in %s", len(targets), targetDir)
	for _, target := range targets {
		log.Debugf("loading target %s", target)
		data, err := ioutil.ReadFile(target)
		if err != nil {
			log.Errorf("error: could not load %s: %v", target, err)
			continue
		}
		targetConf := loadTarget(target, data)
		d.dispatchMap[targetConf.AuthToken] = targetConf
	}
}

// Send formats and sends the message
func (d *Dispatch) Send(request DispatchRequest) error {
	target, found := d.dispatchMap[request.AuthToken]
	if !found {
		return errors.New("auth-token not recognized")
	}

	var email Message
	// if 'from' field is black, email package will fill in a default
	email.FromAddress = target.From
	email.ToAddressList = target.To
	email.Subject = fmt.Sprintf("[dispatch] %s - %s", target.Name, request.Subject)

	var msgBuffer bytes.Buffer
	d.messageTemplate.Execute(&msgBuffer, request)
	email.TextMessage = msgBuffer.String()

	log.Debugf("sending message: %+v", email)
	sendMessage(email, d.smtpSettings)
	return nil
}

// DispatchTarget is a target to send too
type DispatchTarget struct {
	AuthToken string   `yaml:"auth-token"`
	From      string   `yaml:"from"`
	To        []string `yaml:"to"`
	Name      string   `yaml:"name"`
}

func getTargetConfigList(targetDir string) (target []string, err error) {
	pattern := fmt.Sprintf("%s/*", targetDir)
	log.Debugf("searching %s", pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}
	return matches, nil
}

func loadTarget(target string, data []byte) DispatchTarget {
	t := DispatchTarget{}
	err := yaml.Unmarshal(data, &t)
	if err != nil {
		log.Errorf("error: parsing target: %v", err)
	}
	if len(t.Name) == 0 {
		t.Name = path.Base(target)
	}
	log.Debugf("loaded target: %+v", t)
	return t
}
