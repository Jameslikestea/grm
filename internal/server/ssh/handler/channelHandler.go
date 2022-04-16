package handler

import (
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	servicens "github.com/Jameslikestea/grm/internal/namespace/service"
	"github.com/Jameslikestea/grm/internal/policy"
	serviceps "github.com/Jameslikestea/grm/internal/pubkey/service"
	servicers "github.com/Jameslikestea/grm/internal/repository/service"
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

	// Find first occurence of git
	idx := strings.Index(clean, "git")
	if idx == -1 {
		return ""
	}

	return clean[idx:]
}

func cleanRepo(s string) string {
	clean := strings.Map(
		func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.' || r == '-' || r == '/' {
				return r
			}
			return -1
		}, s,
	)

	clean = strings.TrimPrefix(clean, ".")
	clean = strings.TrimPrefix(clean, "/")

	return clean
}

func SSHChannelHandler(ch ssh.Channel, in <-chan *ssh.Request, stor storage.Storage, key string) {
	defer ch.Close()

	var uid = ""

	pol := policy.New()
	ps := serviceps.New(stor)
	rs := servicers.New(stor)
	ns := servicens.New(stor)

	user, _ := ps.GetKey(key)
	if user.UID != "" {
		uid = user.UID
	}

	for req := range in {
		payload := cleanCommand(req.Payload)
		switch req.Type {
		case "exec":
			log.Debug().Str("payload", payload).Msg("Handling Connection")

			cmd := strings.Split(payload, " ")
			if len(cmd) > 1 {
				target := cleanRepo(strings.Join(cmd[1:], " "))

				check := strings.TrimSuffix(target, ".git")
				repo := strings.Split(check, "/")

				if len(repo) != 2 {
					ch.Stderr().Write([]byte("Invalid Repo\n"))
					ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					ch.Close()
					break
				}
				log.Info().Str("namespace", repo[0]).Str("repo", repo[1]).Msg("SSH Repository")

				permissions := rs.GetRepoPermissions(repo[0], repo[1])
				nspermissions := ns.GetNamespacePermissions(repo[0])

				r, _ := rs.GetRepo(repo[0], repo[1])
				n, _ := ns.GetNamespace(repo[0])

				switch cmd[0] {
				case "git-receive-pack":
					if a := pol.Evaluate(
						policy.RepoWrite, policy.PolicyRequest{
							UserID:               uid,
							RepoPermissions:      permissions,
							Repo:                 r,
							NamespacePermissions: nspermissions,
							Namespace:            n,
						},
					); !a {
						ch.Stderr().Write([]byte("Forbidden\n"))
						ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						ch.Close()
						break
					}
					receive.SSHReceivePack(ch, target, stor)
				case "git-upload-pack":
					if a := pol.Evaluate(
						policy.RepoRead, policy.PolicyRequest{
							UserID:               uid,
							RepoPermissions:      permissions,
							Repo:                 r,
							NamespacePermissions: nspermissions,
							Namespace:            n,
						},
					); !a {
						ch.Stderr().Write([]byte("Forbidden\n"))
						ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						ch.Close()
						break
					}
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
