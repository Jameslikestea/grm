package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("grmpkg")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/grmpkg/")
	viper.AddConfigPath("$HOME/.grmpkg/")
	viper.AddConfigPath(".")
}

func SetDefaults() {
	viper.SetDefault(domain, "grmpkg.com")

	viper.SetDefault(sshPort, "2222")
	viper.SetDefault(httpPort, "8080")
	viper.SetDefault(sshInterface, "0.0.0.0")
	viper.SetDefault(httpInterface, "0.0.0.0")

	viper.SetDefault(sshKeyPath, "/etc/grmpkg/grmpkg.rsa")
	viper.SetDefault(sshUser, "git")

	viper.SetDefault(logPath, "/var/log/grmpkg.log")
	viper.SetDefault(logLevel, "INFO")
	viper.SetDefault(logFile, true)
}

func GetConfig() {
	viper.ReadInConfig()
}

func GetConfigFromFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Error().Err(err).Str("path", file).Msg("Config file error")
		GetConfig()
	}
	viper.ReadConfig(f)
}

func WriteConfig() {
	if err := viper.WriteConfig(); err != nil {
		log.Error().Err(err).Msg("Could not write existing config")
		if err = viper.WriteConfigAs("./grmpkg.yml"); err != nil {
			log.Error().Err(err).Msg("Could not write config to current directory")
		}
	}
}
