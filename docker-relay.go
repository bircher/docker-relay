package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {

	// Get the name of the command we are running.
	arg0 := filepath.Base(os.Args[0])

	// Get the arguments we want to pass to the docker command.
	args, err := dockerArgs(arg0, isTTY(), containerID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Find the docker binary.
	binary, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "The "+args[0]+" binary was not found.")
		os.Exit(1)
	}

	// Call to docker and pass on the current environment.
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)

	if execErr != nil {
		panic(execErr)
	}
}

func dockerArgs(arg0 string, tty bool, containerFinder func(string) (string, bool)) ([]string, error) {
	vc, err := viperSub(arg0)
	if err != nil {
		return []string{}, err
	}

	// Try loading the container id from the configuration or use the command name.
	cid := vc.GetString("container")
	if cid == "" {
		cid = arg0
	}

	// Find the docker-compose container with the container id.
	container, ok := containerFinder(cid)
	if !ok {
		return []string{}, fmt.Errorf("The container %s is not running", cid)
	}

	// Parse the viper configuration to a config struct.
	conf, err := getConf(vc, container == "")
	if err != nil {
		return []string{}, fmt.Errorf("Error reading configuration: %e", err)
	}

	// Evaluate environment variables in the config.
	processEnvVar(conf)

	if conf.Exec != nil {
		args := append(conf.Exec, processedArgs(conf)...)
		logDebug(args, conf)
		return args, nil
	}

	// Get a slice of command arguments, if it is just a string it will use that.
	cmd := conf.Cmd

	// Start assembling the command to run.
	args := []string{"docker"}

	// If there is no container to execute then we run the image.
	if container == "" {
		if conf.Image == "" {
			return []string{}, fmt.Errorf("Not using docker-compose but no image is configured")
		}
		args = append(args, "run")
		container = conf.Image
	} else {
		args = append(args, "exec")
	}

	// Always run as interactive.
	args = append(args, "-i")

	// If the program is run as a tty, pass the -t flag to docker.
	if tty {
		args = append(args, "-t")
	}

	// Add the docker arguments based on the configuration.
	args = append(args, conf.options()...)

	others := processedArgs(conf)

	// Add the container, the command and all the arguments passed to us.
	args = append(args, container)
	if cmd != nil {
		args = append(args, cmd...)
	}
	args = append(args, others...)

	logDebug(args, conf)

	return args, nil
}
