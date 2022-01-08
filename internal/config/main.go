package config

import "github.com/spf13/viper"

const (
	domain = "domain"
)

func GetDomain() string {
	return viper.GetString(domain)
}

func SetDomain(d string) {
	viper.Set(domain, d)
}
