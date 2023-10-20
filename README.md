# Docker-relay
This project solves the problem you have when you containerise everything
including the interpreter for your favourite scripting language, yet you still
want to run the scripts from your host. For example if you have a git hook that
executes a php script but php only runs in your docker composed containers.
Also, as a convenience to follow steps in a readme that are not written with
a containerised environment in mind.

## Setup
Clone the repository and run 
`docker run -it --rm -v "$PWD":/usr/src/dr -w /usr/src/dr golang:1.20 go build`
(Or if you have go on your system just run `go build`)
Then symlink docker-relay into your path with the name of the program you want
to relay. For example for php do:
`sudo ln -s ${PWD}/docker-relay /usr/local/bin/php`.

You need docker with a relatively recent version so that it contains compose.
Docker-relay uses `docker compose ps` to find the container. 

## Configuration
docker-relay looks for a `.docker-relay.(yml|toml|json)` file in the current
directory and your home directory. For the sake of this documentation we use
yaml as an example, but you are free to choose your preferred format.

The name of the program/symlink is the top level key.
For example 
```
php:
  container: php # The name of the container in your docker-compose, defaults to the name of the program
  path: "." # The path to replace the current directory with, defaults to empty in which case it is not used.
  cmd: php # The command to run in the container (can be a list)
```

Sometimes a command needs to run inside a docker compose container, and
sometimes it needs to be run in its own container.
For example: `composer create-project`
In case the container can not be found we can run a backup image with
the following configuration:

```
composer:
  container: php # Usually run in the php container
  cmd: composer # run the composer command in it
  user: 1000 # Use user 1000
  run:
    image: 'composer:latest' # use the backup image
    volume: # Mount volumes for composer to interact with the file system
      - '${PWD}:/app'
      - '${HOME}/.cache/composer:/composer'

```

If the container config is set to `!` then it will not look it up and go
straight to running the image.
The configuration under the `run` key inherits everything from the top key
and can override it if needed as well as add new options.
Environment variables are loaded from the `.env` file in the current directory
and can be used if they match the regular expression `\$\{[A-Z_]\}`

## Contributing
We still need unit tests for the go code and documentation of all the options.
Pull requests are welcome.
