package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var version = "v0.5.0"
var dirty = ""

var cfgFile string

var displayVersion string
var showVersion bool
var verbose bool
var debug bool
var check bool

var dispatch *Dispatch

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
	Long:  `Run a webserver that provides an json api for emails`,
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Path to a specific config file (default \"./config.yml\")")
	RootCmd.PersistentFlags().StringP("log-dir", "l", "",
		"Path to log files (default \"/var/log/\")")
	RootCmd.PersistentFlags().StringP("target-dir", "t", "",
		"Path to target configs (default \"/etc/dispatch/targets-enabled\")")
	RootCmd.PersistentFlags().BoolVar(&check, "check", false,
		"Check the config for errors and exit")

	RootCmd.PersistentFlags().BoolVar(&showVersion, "version", false,
		"Display the version number and exit")
	RootCmd.PersistentFlags().StringP("address", "a", "0.0.0.0",
		"The IP address to bind the web server too")
	RootCmd.PersistentFlags().IntP("port", "p", 8080,
		"The port to bind the webserver too")
	RootCmd.PersistentFlags().StringP("rate-limit", "r", "inf",
		"The rate limit at which to send emails in the format 'inf|<num>/<duration>'. "+
			"inf for infinite or 1/10s for 1 email per 10 seconds.")

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
	viper.BindEnv("log_dir")
	viper.BindEnv("target_dir")
	viper.BindEnv("address")
	viper.BindEnv("port")
	viper.BindEnv("rate_limit")
	viper.BindEnv("smtp_server")
	viper.BindEnv("smtp_port")
	viper.BindEnv("smtp_username")
	viper.BindEnv("smtp_password")

	viper.BindPFlag("log_dir", RootCmd.PersistentFlags().Lookup("log-dir"))
	viper.BindPFlag("target_dir", RootCmd.PersistentFlags().Lookup("target-dir"))
	viper.BindPFlag("web.address", RootCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("web.port", RootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("rate-limit", RootCmd.PersistentFlags().Lookup("rate-limit"))
	viper.BindPFlag("smtp.server", RootCmd.PersistentFlags().Lookup("smtp-server"))
	viper.BindPFlag("smtp.port", RootCmd.PersistentFlags().Lookup("smtp-port"))
	viper.BindPFlag("smtp.username", RootCmd.PersistentFlags().Lookup("smtp-username"))
	viper.BindPFlag("smtp.password", RootCmd.PersistentFlags().Lookup("smtp-password"))

	viper.SetDefault("log_dir", "/var/log/")
	viper.SetDefault("target_dir", "/etc/dispatch/targets-enabled")
	viper.SetDefault("web.address", "0.0.0.0")
	viper.SetDefault("web.port", 8080)
	viper.SetDefault("rate-limit", "inf")
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
	} else {
		log.SetLevel(log.InfoLevel)
	}

	logPath := viper.GetString("log_dir")
	logFilePath := path.Join(logPath, "dispatch.log")
	if verbose {
		log.SetOutput(os.Stdout)
		log.Debugf("config: log_dir=%s", logFilePath)
	} else {
		logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file=%v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Infof("config: file=%s", viper.ConfigFileUsed())
	if viper.ConfigFileUsed() == "" {
		log.Fatal("No config file found.")
	}

	smtpSettings := SMTPSettings{
		viper.GetString("smtp.server"),
		viper.GetInt("smtp.port"),
		viper.GetString("smtp.username"),
		viper.GetString("smtp.password"),
	}
	log.Debugf("config: smtp={Host:%s Port:%d UserName:%s}", smtpSettings.Host,
		smtpSettings.Port, smtpSettings.UserName)

	targetsDir := viper.Get("target_dir").(string)
	log.Debugf("config: targets=%s", targetsDir)
	dispatch = NewDispatch(targetsDir, smtpSettings)

	address := viper.GetString("web.address")
	port := viper.GetInt("web.port")

	limitMax, limitTTL, err := getRateLimit(viper.GetString("rate-limit"))
	if err != nil {
		log.Fatalf("error parsing limit: %v", err)
	}

	if check {
		log.Debugf("config: webserver=%s:%d", address, port)
		log.Debugf("config: rate-limit=%d/%s", limitMax, limitTTL)
		log.Infof("Config file format checks out, exiting")
		if !debug {
			log.Infof("Use the --debug flag for more info")
		}
		os.Exit(0)
	}

	// finally, run the webserver
	server := NewServer(dispatch, limitMax, limitTTL)
	server.Run(fmt.Sprintf("%s:%d", address, port))
}

func getRateLimit(rateLimit string) (limitMax int64, limitTTL time.Duration, err error) {
	if rateLimit == "inf" {
		return math.MaxInt64, time.Nanosecond, nil
	}

	parts := strings.Split(rateLimit, "/")
	if len(parts) != 2 {
		msg := fmt.Sprintf("rate limit is not formatted properly - %v", rateLimit)
		return limitMax, limitTTL, errors.New(msg)
	}
	limitMax, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	limitTTL, err = time.ParseDuration(parts[1])
	if err != nil {
		return
	}
	return
}
