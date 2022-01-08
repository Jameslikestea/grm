package config

import "github.com/spf13/viper"

const (
	httpInterface = "http.interface"
	httpPort      = "http.port"
)

func GetHTTPInterface() string {
	return viper.GetString(httpInterface)
}

func SetHTTPInterface(i string) {
	viper.Set(httpInterface, i)
}

func GetHTTPPort() string {
	return viper.GetString(httpPort)
}

func SetHTTPPort(p string) {
	viper.Set(httpPort, p)
}
