package config

import "github.com/spf13/viper"

const (
	sshInterface = "ssh.interface"
	sshPort      = "ssh.port"
	sshKeyPath   = "ssh.keyPath"
	sshUser      = "ssh.username"
)

func GetSSHInterface() string {
	return viper.GetString(sshInterface)
}

func SetSSHInterface(i string) {
	viper.Set(sshInterface, i)
}

func GetSSHPort() string {
	return viper.GetString(sshPort)
}

func SetSSHPort(p string) {
	viper.Set(sshPort, p)
}

func GetSSHKeyPath() string {
	return viper.GetString(sshKeyPath)
}

func SetSSHKeyPath(k string) {
	viper.Set(sshKeyPath, k)
}

func GetSSHUsername() string {
	return viper.GetString(sshUser)
}

func SetSSHUsername(u string) {
	viper.Set(sshUser, u)
}
