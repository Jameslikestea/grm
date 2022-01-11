package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/server/ssh/handler"
	"github.com/Jameslikestea/grm/internal/storage"
	"github.com/Jameslikestea/grm/internal/storage/cql"
	"github.com/Jameslikestea/grm/internal/storage/memory"
)

type Server struct {
	c    *ssh.ServerConfig
	pkey ssh.Signer
	stor storage.Storage
}

func NewServer() *Server {
	pkey, err := ioutil.ReadFile(config.GetSSHKeyPath())
	if err != nil {
		log.Error().Err(err).Msg("Cannot Load SSH Private Key")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return nil
	}

	skey, err := ssh.ParsePrivateKey(pkey)
	if err != nil {
		log.Error().Err(err).Msg("SSH Key provided is not valid")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return nil
	}

	conf := &ssh.ServerConfig{

		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() != config.GetSSHUsername() {
				return nil, ssh.ServerAuthError{
					Errors: []error{
						fmt.Errorf("Username Not Supported. Please use: %s", config.GetSSHUsername()),
					},
				}
			}

			kid := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
			log.Debug().Str("key-id", kid).Str("user", conn.User()).Msg("Received SSH Connection")
			return &ssh.Permissions{Extensions: map[string]string{"key-id": strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))}}, nil
		},
		ServerVersion: "SSH-2.0-GRM",
		BannerCallback: func(conn ssh.ConnMetadata) string {
			return fmt.Sprintf(
				`This server is for Authorized Git Usage only.
By Connecting as %s, you agree that you
are an authorized user, and that your actions
may be audited. Unauthorized connections may
be prosecuted to the full extent of the law.


`, conn.User(),
			)
		},
	}

	conf.AddHostKey(skey)

	var stor storage.Storage

	switch strings.ToUpper(config.GetStorageType()) {
	case "MEMORY":
		stor = memory.NewMemoryStorage()
	// case "S3":
	// 	stor = s3.NewS3Storage()
	case "CQL":
		stor = cql.NewCQLStorage()
	default:
		log.Warn().Msg("No Acceptable Storage Engine Chosen, Defaulting to In Memory")
		stor = memory.NewMemoryStorage()
	}

	return &Server{
		c:    conf,
		pkey: skey,
		stor: stor,
	}
}

func (s *Server) listen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.GetSSHInterface(), config.GetSSHPort()))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start SSH listener")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warn().Err(err).Msg("Error accepting connection")
			continue
		}

		sConn, chans, reqs, err := ssh.NewServerConn(conn, s.c)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot establish SSH connection")
			continue
		}

		go ssh.DiscardRequests(reqs)
		go s.handleConnection(sConn.Permissions.Extensions["key-id"], chans)
	}
}

func (s *Server) handleConnection(keyID string, chans <-chan ssh.NewChannel) {
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
		}

		ch, reqs, err := newChan.Accept()
		if err != nil {
			log.Warn().Err(err).Msg("Cannot accept SSH channel")
		}
		go handler.SSHChannelHandler(ch, reqs, s.stor)
	}
}

func (s *Server) Run() {
	log.Info().Msg("Starting SSH")
	s.listen()
}
