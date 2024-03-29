package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// DispatchMap is a AuthToken to DispatchTarget map
type DispatchMap map[string]DispatchTarget

// DispatchRequest is the expected message submission
type DispatchRequest map[string]string

// DispatchRequest is values provided in headers
type DispatchHeaders map[string]string

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
{{ printf "%-12s" "Timestamp:"}}{{ index . "timestamp" }}
{{ range $key, $value := . -}}
{{ if eq $key "message" "auth-token" "timestamp" }}{{ else -}}
{{title $key | printf "%s:" | printf "%-12s"}}{{$value}}
{{ end }}{{ end -}}
-----------------------------------------------------------
{{ index . "message"}}
`

	d.messageTemplate = template.Must(template.New("request").Funcs(sprig.TxtFuncMap()).Parse(msg))
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
		targetConf, err := loadTarget(target, data)
		if err != nil {
			log.Errorf("error: parsing target %s: %v", target, err)
			continue
		}

		if len(targetConf.To) == 0 {
			log.Errorf("error: target %s does not have a destination, skipping", targetConf.Name)
			continue
		}

		d.dispatchMap[targetConf.AuthToken] = targetConf
		log.Infof("loaded target %s:%s", targetConf.Name, targetConf.AuthToken)
	}
}

// AddTarget adds a target to the dispatch map
func (d *Dispatch) AddTarget(target DispatchTarget) {
	d.dispatchMap[target.AuthToken] = target
}

// Send formats and sends the message
func (d *Dispatch) Send(request DispatchRequest) error {
	target, found := d.dispatchMap[request["auth-token"]]
	if !found {
		return errors.New("authentication is not valid")
	}

	r := mergeRequests(request, target.Defaults)

	// format the email subject line
	subject := ""
	if s, ok := r["subject"]; ok && len(r["subject"]) > 0 {
		subject = fmt.Sprintf(" - %s", s)
	}

	var email Message
	// if 'from' field is black, email package will fill in a default
	email.FromAddress = target.From
	email.ToAddressList = target.To
	email.Subject = fmt.Sprintf("[dispatch] %s%s", target.Name, subject)

	var msgBuffer bytes.Buffer
	d.messageTemplate.Execute(&msgBuffer, request)
	email.TextMessage = msgBuffer.String()

	log.Infof("sending message: {AuthToken:%s Name:%s}", request["auth-token"], request["name"])
	sendMessage(email, d.smtpSettings)
	return nil
}

// DispatchTarget is a target to send too
type DispatchTarget struct {
	AuthToken string            `yaml:"auth-token"`
	From      string            `yaml:"from"`
	To        []string          `yaml:"to"`
	Name      string            `yaml:"name"`
	Defaults  map[string]string `yaml:"defaults"`
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

func loadTarget(target string, data []byte) (DispatchTarget, error) {
	t := DispatchTarget{}
	err := yaml.Unmarshal(data, &t)
	if err != nil {
		return t, err
	}
	if len(t.Name) == 0 {
		t.Name = path.Base(target)
	}

	oldTo := make([]string, len(t.To))
	copy(oldTo, t.To)
	t.To = t.To[:0] // Clear our slice
	for _, addr := range oldTo {
		fAddr, err := FormatEmail(addr)
		if err != nil {
			log.Errorf("error parsing email '%s', skipping", addr)
			continue
		}
		t.To = append(t.To, fAddr)
	}

	log.Debugf("target=%+v", t)
	return t, nil
}

// mergeRequests merges two requests, primary overrides secondary
func mergeRequests(primary, secondary interface{}) map[string]string {
	// primary values always override secondary
	result := map[string]string{}

	vs := reflect.ValueOf(secondary)
	for _, key := range vs.MapKeys() {
		strct := vs.MapIndex(key)
		result[key.Interface().(string)] = strct.Interface().(string)
	}

	vp := reflect.ValueOf(primary)
	for _, key := range vp.MapKeys() {
		strct := vp.MapIndex(key)
		result[key.Interface().(string)] = strct.Interface().(string)
	}
	return result
}
