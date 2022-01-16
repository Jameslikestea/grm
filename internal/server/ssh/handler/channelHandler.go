package handler

import (
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/server/ssh/receive"
	"github.com/Jameslikestea/grm/internal/server/ssh/upload"
	"github.com/Jameslikestea/grm/internal/storage"
)

func cleanCommand(b []byte) string {
	p := string(b)
	clean := strings.Map(
		func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, p,
	)
	return clean
}

func cleanRepo(s string) string {
	clean := strings.Map(
		func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.' || r == '-' {
				return r
			}
			return -1
		}, s,
	)

	return clean
}

func SSHChannelHandler(ch ssh.Channel, in <-chan *ssh.Request, stor storage.Storage) {
	defer ch.Close()

	for req := range in {
		payload := cleanCommand(req.Payload)
		switch req.Type {
		case "exec":
			log.Debug().Str("payload", payload).Msg("Handling Connection")

			cmd := strings.Split(payload, " ")
			if len(cmd) > 1 {
				target := cleanRepo(strings.Join(cmd[1:], " "))
				switch cmd[0] {
				case "git-receive-pack":
					receive.SSHReceivePack(ch, target, stor)
				case "git-upload-pack":
					upload.SSHUploadPack(ch, target, stor)
				default:
				}
			}

			ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			ch.Close()
			break
		default:
		}

	}
}
