package main

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/gin-gonic/gin.v1"

	log "github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var version = "v0.1.0"
var dirty = ""

var cfgFile string
var logPath string

var displayVersion string
var showVersion bool
var verbose bool
var debug bool
var check bool

var dispatchMap map[string]dispatch
var smtpSettings SMTPSettings

func main() {
	displayVersion = fmt.Sprintf("dispatch %s%s",
		version,
		dirty)
	Execute(displayVersion)
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "A mail forwarding API service",
	Long:  `This app provides an json-api for sending email`,
	Run:   run,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	displayVersion = version
	RootCmd.SetHelpTemplate(fmt.Sprintf("%s\nVersion:\n  github.com/gesquive/%s\n",
		RootCmd.HelpTemplate(), displayVersion))
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	//TODO: add --check flag to check for config corruption
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Path to a specific config file (default \"./config.yaml\")")
	RootCmd.PersistentFlags().String("log-path", "",
		"Path to log files (default \"/var/log/\")")
	RootCmd.PersistentFlags().BoolVar(&check, "check", false,
		"Check the config for errors and exit")

	RootCmd.PersistentFlags().BoolVar(&showVersion, "version", false,
		"Display the version number and exit")
	RootCmd.PersistentFlags().StringP("address", "a", "0.0.0.0",
		"The IP address to bind the web server too")
	RootCmd.PersistentFlags().IntP("port", "p", 8080,
		"The port to bind the webserver too")

	RootCmd.PersistentFlags().StringP("smtp-server", "x", "localhost",
		"The SMTP server to send email through")
	RootCmd.PersistentFlags().Uint32P("smtp-port", "o", 25,
		"The port to use for the SMTP server")
	RootCmd.PersistentFlags().StringP("smtp-username", "u", "",
		"Authenticate the SMTP server with this user")
	RootCmd.PersistentFlags().StringP("smtp-password", "w", "",
		"Authenticate the SMTP server with this password")

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Print logs to stdout instead of file")

	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false,
		"Include debug statements in log output")
	RootCmd.PersistentFlags().MarkHidden("debug")

	viper.SetEnvPrefix("dispatch")
	viper.AutomaticEnv()
	viper.BindEnv("address")
	viper.BindEnv("port")
	viper.BindEnv("smtp_server")
	viper.BindEnv("smtp_port")
	viper.BindEnv("smtp_username")
	viper.BindEnv("smtp_password")

	viper.BindPFlag("log_path", RootCmd.PersistentFlags().Lookup("log-path"))
	viper.BindPFlag("web.address", RootCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("web.port", RootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("smtp.server", RootCmd.PersistentFlags().Lookup("smtp-server"))
	viper.BindPFlag("smtp.port", RootCmd.PersistentFlags().Lookup("smtp-port"))
	viper.BindPFlag("smtp.username", RootCmd.PersistentFlags().Lookup("smtp-username"))
	viper.BindPFlag("smtp.password", RootCmd.PersistentFlags().Lookup("smtp-password"))

	viper.SetDefault("log_path", "/var/log/")
	viper.SetDefault("web.address", "0.0.0.0")
	viper.SetDefault("web.port", 8080)
	viper.SetDefault("smtp.server", "localhost")
	viper.SetDefault("smtp.port", 25)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("config")                 // name of config file (without extension)
	viper.AddConfigPath(".")                      // add current directory as first search path
	viper.AddConfigPath("$HOME/.config/dispatch") // add home directory to search path
	viper.AddConfigPath("/etc/dispatch")          // add etc to search path
	viper.AutomaticEnv()                          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if !showVersion {
			log.Error("Error opening config: ", err)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	if showVersion {
		fmt.Println(displayVersion)
		os.Exit(0)
	}

	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: time.RFC3339,
	})

	if debug {
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	logPath = path.Dir(viper.GetString("log_path"))
	logPath = fmt.Sprintf("%s/dispatch.log", logPath)
	if verbose {
		log.SetOutput(os.Stdout)
		log.Debugf("config: log_file=%s", logPath)
	} else {
		logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file=%v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Debugf("config: file=%s", viper.ConfigFileUsed())
	if viper.ConfigFileUsed() == "" {
		log.Fatal("No config file found.")
	}

	dispatchMap = getDispatchMap(viper.Get("dispatch"))
	for _, val := range dispatchMap {
		log.Debugf("config: dispatch=%+v", val)
	}
	smtpSettings = SMTPSettings{
		viper.GetString("smtp.server"),
		viper.GetInt("smtp.port"),
		viper.GetString("smtp.username"),
		viper.GetString("smtp.password"),
	}
	log.Debugf("config: smtp={Host:%s Port:%d UserName:%s}", smtpSettings.Host,
		smtpSettings.Port, smtpSettings.UserName)

	address := viper.GetString("web.address")
	port := viper.GetInt("web.port")
	if check {
		log.Debugf("config: webserver=%s:%d", address, port)
		log.Infof("Config file format checks out, exiting")
		if !debug {
			log.Infof("Use the --debug flag for more info")
		}
		os.Exit(0)
	}

	// finally, run the webserver
	router := gin.New()
	router.Use(logger)

	router.POST("/send", send)

	log.Infof("Starting webserver on %s:%d", address, port)
	router.Run(fmt.Sprintf("%s:%d", address, port))
}

func getDispatchMap(rawMap interface{}) map[string]dispatch {
	s := reflect.ValueOf(rawMap)
	if s.Kind() != reflect.Slice {
		log.Fatal("config: dispatch section is not properly formatted")
	}

	dispatchMap := make(map[string]dispatch)
	for i := 0; i < s.Len(); i++ {
		dmap := s.Index(i).Interface().(map[interface{}]interface{})
		// check for mandatory fields
		authToken, ok := dmap["auth-token"]
		if !ok {
			log.Fatal("config: dispatch.auth-token is missing")
		}
		to, ok := dmap["to"]
		if !ok {
			log.Fatal("config: dispatch.to field is missing")
		}

		var d dispatch
		d.AuthToken = authToken.(string)
		if from, ok := dmap["from"]; ok {
			d.From = from.(string)
		}

		if reflect.ValueOf(to).Kind() == reflect.String {
			d.To = []string{dmap["to"].(string)}
		} else if reflect.ValueOf(to).Kind() == reflect.Slice {
			d.To = make([]string, len(dmap["to"].([]interface{})))
			for i, a := range dmap["to"].([]interface{}) {
				d.To[i] = a.(string)
			}
		} else {
			log.Fatal("config: dispatch.to field is not properly formatted")
		}

		dispatchMap[d.AuthToken] = d
	}
	return dispatchMap
}

func logger(c *gin.Context) {
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
