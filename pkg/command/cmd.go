package command

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"runtime"
)

type Cmd struct {
	cmd    *exec.Cmd
	Stdin  io.WriteCloser
	stdout io.ReadCloser
}

func NewCmd() *Cmd {
	var shell string
	switch os := runtime.GOOS; os {
	case "darwin":
		shell = "/bin/zsh"
	case "linux":
		shell = "/bin/bash"
	case "windows":
		shell = "cmd.exe"
	default:
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Stderr = cmd.Stdout

	return &Cmd{cmd, stdin, stdout}
}

func (c *Cmd) Outpout(out chan []byte) {
	go func() {
		for {
			buff := bufio.NewReader(c.stdout)
			for {
				outText, err := buff.ReadBytes('\n')
				if err != nil {
					log.Fatal(err)
				}
				out <- outText
			}
		}
	}()

	// if err := c.cmd.Start(); err != nil {
	// 	log.Fatalf("Failed to start cmd: %v\n", err)

	// }
}
func (c *Cmd) Run() {
	c.cmd.Run()
}
