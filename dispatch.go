package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// DispatchMap is a AuthToken to DispatchTarget map
type DispatchMap map[string]DispatchTarget

// DispatchRequest is the expected message submission
type DispatchRequest struct {
	AuthToken string `json:"auth-token" binding:"required"`
	Message   string `json:"message"`
	Subject   string `json:"subject"`
}

// Dispatch is the central point for the dispatches
type Dispatch struct {
	dispatchMap  DispatchMap
	smtpSettings SMTPSettings
}

// NewDispatch create a new dispatch
func NewDispatch(targetDir string, smtpSettings SMTPSettings) *Dispatch {
	d := new(Dispatch)
	d.dispatchMap = make(DispatchMap)
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
		targetConf := loadTarget(data)
		d.dispatchMap[targetConf.AuthToken] = targetConf
	}
}

// Send places the request in the queue
func (d *Dispatch) Send(request DispatchRequest) error {
	target, found := d.dispatchMap[request.AuthToken]
	if !found {
		return errors.New("auth-token not recognized")
	}
	//TODO: Rate-limit by ip address here

	var email Message
	// if 'from' field is black, email package will fill in a default
	email.FromAddress = target.From
	email.ToAddressList = target.To
	if len(target.SubjectPrefix) > 0 {
		email.Subject = fmt.Sprintf("%s %s", target.SubjectPrefix, request.Subject)
	} else {
		email.Subject = request.Subject
	}
	if len(target.MessagePrefix) > 0 {
		email.TextMessage = fmt.Sprintf("%s\n%s", target.MessagePrefix, request.Message)
	} else {
		email.TextMessage = request.Message
	}

	log.Debugf("sending message: %+v", email)
	sendMessage(email, d.smtpSettings)
	return nil
}

// DispatchTarget is a target to send too
type DispatchTarget struct {
	AuthToken     string   `yaml:"auth-token"`
	From          string   `yaml:"from"`
	To            []string `yaml:"to"`
	SubjectPrefix string   `yaml:"subject-prefix"`
	Subject       string   `yaml:"subject"`
	MessagePrefix string   `yaml:"message-prefix"`
}

func getTargetConfigList(targetDir string) (target []string, err error) {
	pattern := fmt.Sprintf("%s/*.yml", targetDir)
	log.Debugf("searching %s", pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}
	return matches, nil
}

func loadTarget(data []byte) DispatchTarget {
	t := DispatchTarget{}
	err := yaml.Unmarshal(data, &t)
	if err != nil {
		log.Errorf("error: parsing target: %v", err)
	}
	log.Debugf("loaded target: %+v", t)
	return t
}
