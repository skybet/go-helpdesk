// package main is a fully working example implementation of go-helpdesk
package main

import (
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/skybet/go-helpdesk/handlers"
	"github.com/skybet/go-helpdesk/server"
	"github.com/skybet/go-helpdesk/wrapper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Generate mocks - go get github.com/vektra/mockery/.../ first
//go:generate mockery -name SlackWrapper -recursive

func main() {
	initFlags()
	// Connect to Slack
	appToken := viper.GetString("app-token")
	botToken := viper.GetString("bot-token")
	signingSecret := viper.GetString("signing-secret")
	if appToken == "" || botToken == "" || signingSecret == "" {
		pflag.PrintDefaults()
		return
	}
	sw, err := wrapper.New(appToken, botToken)
	if err != nil {
		log.Fatalf("Error initialising the Slack API: %s", err)
	}
	handlers.Init(sw)
	log.Info("Connected to Slack API")
	// Start a server to respond to callbacks from Slack
	s := server.NewSlackHandler("/slack", appToken, signingSecret, nil, log.Info, log.Infof, log.Error, log.Errorf)
	s.HandleCommand("/help-me", handlers.HelpRequest)
	s.HandleInteractionCallback("dialog_submission", "HelpRequest", handlers.HelpCallback)
	addr := viper.GetString("listen-address")
	go func() {
		if err := http.ListenAndServe(addr, s); err != nil {
			log.Fatalf("Unable to start server: %s", err)
		}
	}()
	log.Infof("Listening for Slack callbacks on '%s'", addr)
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	<-terminate
}

func initFlags() {
	// Bind flags
	pflag.StringP("app-token", "a", "", "Slack API token for your slash command (required)")
	pflag.StringP("bot-token", "b", "", "Slack API token for bot integration (required)")
	pflag.StringP("signing-secret", "s", "", "Slack API signing secret for request verification (required)")
	pflag.StringP("listen-address", "l", ":4390", "Address to listen for Slack callbacks on")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	// Allow setting flags from environment variables
	// An explicitly set flag will take precedence over the environment variable
	// Example: HELP_APP_TOKEN will set the app-token flag
	viper.SetEnvPrefix("help")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}
