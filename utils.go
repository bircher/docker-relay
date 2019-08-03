package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {

	// Load the environment variables from the .env file.
	// This makes them available to be processed in the config file.
	// The variables still needs to be mapped with --env in order to be available in the container.
	// Unfortunately we can not use environment variables to override
	// the viper configuration because we use sub configuration and marshalling.
	godotenv.Load()

	// Look for a .docker-relay.[toml|yml|json] file.
	filename := ".docker-relay"

	// Look for .dockwrap.yml config file
	viper.SetConfigName(filename)

	// Use the home directory if available.
	if home, err := homedir.Dir(); err == nil {
		_ = home
		viper.AddConfigPath(home)
		viper.ReadInConfig()
	}

	// Merge in local configuration.
	if wd, err := os.Getwd(); err == nil {
		// Check for files in the current directory and merge then in if they exist.
		for _, ext := range []string{".json", ".yaml", ".yml", ".toml"} {
			fn := wd + "/" + filename + ext
			if _, err := os.Stat(fn); err == nil {
				viper.SetConfigFile(fn)
				viper.MergeInConfig()
			}

		}

	}
}

// viperSub returns the viper config for a key, and makes sure it it not nil.
func viperSub(name string) (*viper.Viper, error) {
	v := viper.Sub(name)
	if v == nil {
		return nil, fmt.Errorf("The configuration for %s does not exist", name)
	}
	return v, nil
}

// containerID returns the container id or an empty string and ok
func containerID(name string) (string, bool) {
	if name == "!" {
		// If the name of the container is a ! then we want to skip looking it up.
		return "", true
	}

	out, err := exec.Command("docker-compose", "ps", "-q", name).Output()
	if err != nil {
		// If docker-compose complains about not finding the container, it is ok and we want to run the image.
		return "", true
	}

	id := strings.TrimSpace(string(out))

	return id, id != ""
}

// isTTY returns true if the current program runs as a tty.
func isTTY() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
}

// processEnvVar escapes $(pwd) and environment variables
func processEnvVar(c *config) {

	fp, err := os.Getwd()
	if err == nil {
		pwdr := regexp.MustCompile(`\$\{PWD\}|\$PWD|\$\(pwd\)`)
		trans := func(in string, n string) string {
			return fp
		}
		c.replaceStrings(pwdr, trans)
	}

	r := regexp.MustCompile(`\$\{[A-Z_]+\}`)
	f := func(in string, n string) string {
		v := strings.Replace(strings.Replace(in, "${", "", 1), "}", "", 1)
		if env, ok := os.LookupEnv(v); ok {
			return env
		}
		return in
	}

	c.replaceStrings(r, f)
}

// processedArgs gets the arguments passed to the program and processes absolute paths.
func processedArgs(c *config) []string {

	args := os.Args[1:]
	// When wrapping a scripting language like php, often the script might
	// contain the absolute path, if the configuration has a path use that.
	if c.Path != "" {
		fp, err := os.Getwd()
		if err == nil {
			for i, el := range args {
				args[i] = strings.Replace(el, fp, c.Path, 1)
			}
		}
	}

	return args
}

// Debug checks the config if debugging is enabled and logs the command about to run to a file.
func logDebug(cmd []string, cfg *config) {
	// Debug output of the command which is about to run.
	if viper.GetBool("docker-relay-debug.enabled") {
		fn := viper.GetString("docker-relay-debug.file")
		if fn == "" {
			fn = "docker-relay-debug.txt"
		}
		f, err := os.Create(fn)
		defer f.Close()
		if err == nil {
			fp, err := os.Getwd()
			if err == nil {
				fmt.Fprintln(f, fp)
			}
			fmt.Fprintln(f, "---")
			fmt.Fprintln(f, cmd)
			fmt.Fprintln(f, "---")
			fmt.Fprintln(f, cfg)
			fmt.Fprintln(f, "---")
			fmt.Fprintln(f, cfg.options())

		}
	}
}
