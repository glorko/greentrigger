package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ServiceProcess struct {
	Command *exec.Cmd
	PID     int
	Port    int            // Dynamically generated port
	Config  *ServiceConfig // Link to the configuration
}

// StartServer starts the service and monitors its logs for a startup message.
func StartServer(service *ServiceConfig) (*ServiceProcess, error) {
	// Hardcoded port for now
	port := 8080

	// Start the external process and redirect logs
	cmd, stdoutReader, stderrReader, err := startProcess(service, port)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Service %s started with PID %d on port %d\n", service.Name, cmd.Process.Pid, port)

	// Monitor the logs for startup completion
	err = monitorLogs(service.Name, stdoutReader, stderrReader)
	if err != nil {
		return nil, err
	}

	// Return the service process details
	return &ServiceProcess{
		Command: cmd,
		PID:     cmd.Process.Pid,
		Port:    port,
		Config:  service,
	}, nil
}

// startProcess starts the service process and sets up stdout and stderr piping.
func startProcess(service *ServiceConfig, port int) (*exec.Cmd, *bufio.Reader, *bufio.Reader, error) {
	var cmd *exec.Cmd

	// Construct the command based on the platform
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("%s --port %d", service.Command, port))
	} else {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("%s --port %d", service.Command, port))
	}

	// Inherit the environment from the parent process (including PATH)
	cmd.Env = os.Environ() // This copies the current environment variables, including PATH

	// Add additional environment variables if needed
	cmd.Env = append(cmd.Env, fmt.Sprintf("GLOBAL_VAR=%s", service.Env["GLOBAL_VAR"]))

	// Log the full command for debugging
	fmt.Printf("Executing command: %s\n", cmd.String())

	// Pipe stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to get stdout pipe for service %s: %v", service.Name, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to get stderr pipe for service %s: %v", service.Name, err)
	}

	// Start the external process
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to start service %s: %v", service.Name, err)
	}

	return cmd, bufio.NewReader(stdout), bufio.NewReader(stderr), nil
}

// monitorLogs reads stdout and stderr and checks for the startup completion message.
func monitorLogs(serviceName string, stdoutReader *bufio.Reader, stderrReader *bufio.Reader) error {
	startedChan := make(chan bool)

	// Goroutine to read stdout and look for "Startup completed"
	go func() {
		for {
			line, err := stdoutReader.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Printf("[LOG - %s]: %s", serviceName, line)

			// Detect "Startup completed" log and extract startup time
			if strings.Contains(line, "Startup completed") {
				startedChan <- true
				var startupTime string
				fmt.Sscanf(line, "Startup completed in %s", &startupTime)
				fmt.Printf("Service %s startup time parsed from logs: %s\n", serviceName, startupTime)
			}
		}
	}()

	// Goroutine to read stderr logs
	go func() {
		for {
			line, err := stderrReader.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Printf("[ERROR - %s]: %s", serviceName, line)
		}
	}()

	// Wait for "Startup completed" signal or timeout
	select {
	case <-startedChan:
		fmt.Printf("Service %s fully started.\n", serviceName)
	case <-time.After(30 * time.Second): // Adjust timeout as needed
		fmt.Printf("Service %s startup timeout.\n", serviceName)
		return fmt.Errorf("Service %s failed to start within timeout", serviceName)
	}

	return nil
}
