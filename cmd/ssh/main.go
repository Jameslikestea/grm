package main

import (
	"log"

	"grmpkg-ssh/internal/sshcmd"
)

func main() {
	err := sshcmd.NewCommand().Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
