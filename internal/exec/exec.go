package exec

import (
	"os/exec"
)

func BashCommand(cmd string) (string, error) {
	c := exec.Command("/bin/bash", "-c", cmd)
	out, err := c.Output()
	return string(out), err
}

func BashFile(file string, args ...string) error {
	err := exec.Command("/bin/bash", append([]string{file}, args...)...).Run()
	return err
}
