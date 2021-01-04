package wrapper

import (
	"bufio"
	"fmt"
	"os/exec"
)

// console contains the cmd with its IO reader and writer
type console struct {
	cmd    *exec.Cmd
	stdout *bufio.Reader
	stderr *bufio.Reader
	stdin  *bufio.Writer
}

// newConsole initialises a new console
func newConsole(cmd *exec.Cmd) *console {
	c := &console{
		cmd: cmd,
	}

	stdout, _ := cmd.StdoutPipe()
	c.stdout = bufio.NewReader(stdout)

	stderr, _ := cmd.StderrPipe()
	c.stderr = bufio.NewReader(stderr)

	stdin, _ := cmd.StdinPipe()
	c.stdin = bufio.NewWriter(stdin)

	return c
}

// Start starts the console
func (c *console) Start() error {
	return c.cmd.Start()
}

// WriteCmd writes to the console
func (c *console) WriteCmd(cmd string) error {
	wrappedCmd := fmt.Sprintf("%s\n", cmd)
	_, err := c.stdin.WriteString(wrappedCmd)
	if err != nil {
		return err
	}
	return c.stdin.Flush()
}

// ReadLine reads a line from sdtout
func (c *console) ReadLine() (string, error) {
	return c.stdout.ReadString('\n')
}

// ReadErr reads a line from sdterr
func (c *console) ReadErr() (string, error) {
	return c.stderr.ReadString('\n')
}

// Kill kills the Consoleout
func (c *console) Kill() error {
	return c.cmd.Process.Kill()
}
