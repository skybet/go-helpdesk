// package main is a fully working example implementation of go-helpdesk
package main

import (
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

func main() {
	initFlags()
	// Connect to Slack
	appToken := viper.GetString("app-token")
	botToken := viper.GetString("bot-token")
	if appToken == "" || botToken == "" {
		pflag.PrintDefaults()
		return
	}
	if err := wrapper.Init(appToken, botToken); err != nil {
		log.Fatalf("Error initialising the Slack API: %s", err)
	}
	log.Info("Connected to Slack API")
	// Start a server to respond to callbacks from Slack
	s := server.NewSlackReceiver()
	sc := &server.SlashCommand{
		Path:    "/slack/command/help",
		Handler: handlers.DialogTest,
	}

	if err := s.AddSlashCommand(sc); err != nil {
		log.Fatalf("Failed to add slash command to server: %s", err)
	}
	addr := viper.GetString("listen-address")
	go func() {
		if err := s.Start(addr, log.Errorf); err != nil {
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
