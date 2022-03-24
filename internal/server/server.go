package server

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/dateiexplorer/minecrafter/internal/exec"
	"github.com/mikroskeem/mcping"
)

type ServerStatus string

const (
	Locked    ServerStatus = "LOCKED"
	Stopping  ServerStatus = "STOPPING"
	Down      ServerStatus = "DOWN"
	Starting  ServerStatus = "STARTING"
	Up        ServerStatus = "UP"
	Undefined ServerStatus = "UNDEFINED"
)

type Server struct {
	base string
	name string
	ip   string
}

func FromServerName(base string, name string, ip string) (*Server, error) {
	// Get the real name of the server. This ensures that 'current' points to
	// the real server name.
	cmd := fmt.Sprintf("echo -n $(basename $(realpath '%v/%v'))", base, name)
	realName, err := exec.BashCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("cannot get server from name '%v'", name)
	}

	s := &Server{base, realName, ip}
	return s, nil
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) IP() string {
	return s.ip
}

func (s *Server) Status() (ServerStatus, *mcping.PingResponse, error) {
	// Check if the server is stopping currently
	stopping, err := s.isStopping()
	if err != nil {
		return Undefined, nil, err
	}
	if stopping {
		// A stop screen is running
		return Stopping, nil, nil
	}

	// Check if a server is running currently
	running, err := s.isRunning()
	if err != nil {
		return Undefined, nil, err
	}
	if running {
		// A start screen is running
		res, err := mcping.Ping(s.ip)
		if err != nil {
			// Server is not ready yet
			return Starting, nil, nil
		}
		// Server is up and ready
		return Up, &res, nil
	}

	// Server is DOWN
	locked, err := s.isLocked()
	if err != nil {
		return Undefined, nil, err
	}
	if locked {
		return Locked, nil, nil
	}

	return Down, nil, nil
}

func (s *Server) Run(pstCLIPath string) error {
	err := exec.BashFile(pstCLIPath, "run", s.name)
	if err != nil {
		return fmt.Errorf("cannot run server '%v': %w", s.name, err)
	}

	return nil
}

func (s *Server) Stop(pstCLIPath string) error {
	err := exec.BashFile(pstCLIPath, "stop", s.name, "now")
	if err != nil {
		return fmt.Errorf("cannot stop server '%v': %w", s.name, err)
	}

	return nil
}

func (s *Server) isStopping() (bool, error) {
	cmd := fmt.Sprintf("echo -n $(screen -ls | grep '%v-stop' | wc -l)", s.name)
	out, err := exec.BashCommand(cmd)
	if err != nil {
		return false, fmt.Errorf("failed to execute command: %w", err)
	}

	count, err := strconv.Atoi(out)
	if err != nil {
		return false, fmt.Errorf("cannot parse command output: %w", err)
	}

	return count > 0, nil
}

func (s *Server) isRunning() (bool, error) {
	cmd := fmt.Sprintf("echo -n $(screen -ls | grep '%v-run' | wc -l)", s.name)
	out, err := exec.BashCommand(cmd)
	if err != nil {
		return false, fmt.Errorf("failed to execute command: %w", err)
	}

	count, err := strconv.Atoi(out)
	if err != nil {
		return false, fmt.Errorf("cannot parse command output: %w", err)
	}

	return count > 0, nil
}

func (s *Server) isLocked() (bool, error) {
	_, err := os.Stat(path.Join(s.base, s.name, "lock"))
	// Error must be os.ErrNotExist, if any other error occurred
	// return the error
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get file information: %w", err)
	}

	return true, nil
}
