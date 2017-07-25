// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func connectSSH(user, host, key string, port int) {
	args := []string{
		"-p", fmt.Sprintf("%d", port),
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=quiet",
		fmt.Sprintf("%s@%s", user, host),
	}

	if key != "" {
		args = append(args, "-i")
		args = append(args, key)
	}

	cmd := exec.Command("ssh", args...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
