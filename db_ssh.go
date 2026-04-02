package grest

import (
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

// StartSSHTunnel starts an SSH tunnel based on the provided SSH configuration.
func StartSSHTunnel(cfg SSHConfig) error {
	key, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.LocalPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				continue
			}

			go func() {
				remoteAddr := fmt.Sprintf("%s:%d", cfg.RemoteHost, cfg.RemotePort)
				remoteConn, err := client.Dial("tcp", remoteAddr)
				if err != nil {
					return
				}

				go io.Copy(remoteConn, localConn)
				io.Copy(localConn, remoteConn)
			}()
		}
	}()

	return nil
}
