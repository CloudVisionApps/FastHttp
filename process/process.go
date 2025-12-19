package process

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/fatih/color"
)

const PIDFile = "/var/run/fasthttp.pid"

func WritePID() error {
	pid := os.Getpid()

	file, err := os.Create(PIDFile)
	if err != nil {
		return fmt.Errorf("error creating PID file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("error writing to PID file: %w", err)
	}

	return nil
}

func ReadPID() (int, error) {
	pidBytes, err := os.ReadFile(PIDFile)
	if err != nil {
		return 0, fmt.Errorf("error reading PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return 0, fmt.Errorf("error converting PID to integer: %w", err)
	}

	return pid, nil
}

func Stop() error {
	log.Println("Shutting down server...")

	pid, err := ReadPID()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %w", err)
	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("error killing process: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

func Status(httpPort string) error {
	pid, err := ReadPID()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %w", err)
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		color.Red("Server is not running")
	} else {
		color.Green("Server is running on port " + httpPort + " with PID " + strconv.Itoa(pid))
	}

	return nil
}
