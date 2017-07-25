// +build windows

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

func connectSSH(user, host, key string, port int) {
	fd := int(os.Stdin.Fd())
	prevState, err := terminal.MakeRaw(fd)
	if err != nil {
		return
	}
	defer terminal.Restore(fd, prevState)

	// vultr ssh needs to be in a terminal
	if terminal.IsTerminal(fd) {
		var config *ssh.ClientConfig

		// use keyfile if provided
		if key != "" {
			signer, err := ssh.ParsePrivateKey(readKeyFile(key))
			if err != nil {
				log.Println("Could not parse private key file")
				log.Fatal(err)
			}
			config = &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
			}
		} else if os.Getenv("SSH_AUTH_SOCK") != "" {
			// if no keyfile is provided, try to use ssh-agent
			conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			sshAgent := agent.NewClient(conn)
			config = &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeysCallback(sshAgent.Signers),
				},
			}
		}

		// connect to ssh
		client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", host, port), config)
		if err != nil {
			if strings.Contains(err.Error(), "no supported methods remain") {
				// let's see if we can maybe give it one more try by using an interactive password prompt
				fmt.Print("Password: ")
				password, err := terminal.ReadPassword(fd)
				if err != nil {
					log.Fatal(err)
				}

				config = &ssh.ClientConfig{
					User: user,
					Auth: []ssh.AuthMethod{
						ssh.Password(string(password)),
					},
				}
				client, err = ssh.Dial("tcp", fmt.Sprintf("%v:%v", host, port), config)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()

		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		session.Stdin = os.Stdin

		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			return
		}

		modes := ssh.TerminalModes{
			ssh.ECHO:          0,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		if err := session.RequestPty("xterm", termWidth, termHeight, modes); err != nil {
			log.Fatal(err)
		}

		if err := session.Shell(); err != nil {
			log.Fatal(err)
		}
		if err := session.Wait(); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("vultr ssh needs an interactive terminal to work")
	}
}

func readKeyFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Could not read ssh key file: " + filename)
		log.Fatal(err)
	}
	return data
}
