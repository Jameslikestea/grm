package handler

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func SSHChannelHandler(ch ssh.Channel, in <-chan *ssh.Request) {
	defer ch.Close()

	for req := range in {
		payload := string(req.Payload)
		switch req.Type {
		case "exec":
			log.Debug().Str("payload", payload).Msg("Handling Connection")
		default:
		}
	}
}
