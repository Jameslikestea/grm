package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	authenticationProvider           = "authentication.provider"
	authenticationGithubClientID     = "authentication.github.clientId"
	authenticationGithubClientSecret = "authentication.github.clientSecret"
	authenticationGithubRedirectURL  = "authentication.github.baseURL"
)

func GetAuthenticationProvider() string {
	return strings.ToUpper(viper.GetString(authenticationProvider))
}

func GetAuthenticationGithubClientID() string {
	return viper.GetString(authenticationGithubClientID)
}

func GetAuthenticationGithubClientSecret() string {
	return viper.GetString(authenticationGithubClientSecret)
}

func GetAuthenticationGithubRedirectURL() string {
	return viper.GetString(authenticationGithubRedirectURL)
}
