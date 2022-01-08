package main

import (
	"log"

	"github.com/Jameslikestea/grm/internal/sshcmd"
)

func main() {
	err := sshcmd.NewCommand().Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
