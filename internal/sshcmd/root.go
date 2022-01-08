package sshcmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/server/http"
	"github.com/Jameslikestea/grm/internal/server/ssh"
)

var rootCmd *cobra.Command

type flag struct {
	Name  string
	Set   func(v string)
	Usage string
}

func NewCommand() *cobra.Command {
	if rootCmd == nil {
		initRootCmd()
	}
	return rootCmd
}

func initRootCmd() {
	f := []flag{
		{
			Name:  "domain",
			Set:   config.SetDomain,
			Usage: "Set the domain base for packages, e.g. grmpkg.com",
		},
		{
			Name:  "http.port",
			Set:   config.SetHTTPPort,
			Usage: "Set the HTTP port (8080)",
		},
		{
			Name:  "http.interface",
			Set:   config.SetHTTPInterface,
			Usage: "Set the HTTP interface (0.0.0.0)",
		},
		{
			Name:  "ssh.port",
			Set:   config.SetSSHPort,
			Usage: "Set the SSH port (8080)",
		},
		{
			Name:  "ssh.interface",
			Set:   config.SetSSHInterface,
			Usage: "Set the SSH interface (0.0.0.0)",
		},
	}

	rootCmd = &cobra.Command{
		Use:   "grmpkg",
		Short: "GRM pkg is a git package manager designed for use with the Go programming language",
		Long:  "GRM pkg is a complete package manager for managing immutable Git tags for use with Go Modules",
		Run: func(cmd *cobra.Command, args []string) {
			config.SetDefaults()
			config.GetConfig()
			config.WriteConfig()

			for _, fl := range f {
				v := rootCmd.Flag(fl.Name)
				if v.Changed {
					fl.Set(v.Value.String())
				}
			}

			zerolog.SetGlobalLevel(config.GetLogLevel())

			if config.GetLogFile() {
				f, err := os.OpenFile(config.GetLogPath(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
				if err != nil {
					log.Panic().Err(err).Msg("Cannot Open Logfile")
					return
				}
				log.Logger = zerolog.New(f).With().Timestamp().Logger()
			}

			log.Info().Msg("Starting Services")

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			h := http.NewServer()
			s := ssh.NewServer()

			go h.Run()
			go s.Run()

			<-c
			log.Info().Msg("Stopping Services")
		},
	}

	for _, fl := range f {
		rootCmd.Flags().String(fl.Name, "", fl.Usage)
	}
}
