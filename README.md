# KEYMAN

## Table of Contents

- [KEYMAN](#keyman)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
  - [Code Overview](#code-overview)
  - [Getting Started](#getting-started)
    - [Dependencies](#dependencies)
      - [For Using the Project](#for-using-the-project)
      - [For Development](#for-development)
    - [Usage](#usage)
    - [Development](#development)
  - [Testing](#testing)
  - [TODO](#todo)

## About

This project is an API for managing keys and secret data nessesary for [RJ's main site](https://therileyjohnson.com)

## Code Overview

The program in `gatekeeper/` acts as a protective reverse proxy for the API in `keymanager/`. The `gatekeeper/` program screens a request, makes sure that it is from an authorized source, rejects if if it is not or forwards it to the API in `keymanager/` if it is.

## Getting Started

### Dependencies

#### For Using the Project

- Docker

- Docker-Compose

#### For Development

- The Dependencies Outlined Above for Using the Project

- Go

- The modules in each project (this will be transformed into a project that uses go modules for dependency management soon)

To install the modules for the `gatekeeper/` program (the process is the same for the `keymanager/` program):

```bash
rj@desktop:~/gatekeeper$ go get ./...
```

### Usage

```bash
rj@desktop:~/gatekeeper$ docker-compose up
```

### Development

Before deciding what you want to change, you first need to clone the project, or fork and branch off of master if you are planning to submit a pull request.

(note, if you forked the project, you will want to change the `/the-rileyj/` portion of the  URL to your github user name)

Installing into GOPATH:

```bash
rj@desktop:~/$ go get github.com/the-rileyj/KeyMan && cd $GOPATH/src/github.com/the-rileyj/KeyMan && git checkout -b <branch-name>
```

Or into the directory of your choosing:

```bash
rj@desktop:~/$ git clone https://github.com/the-rileyj/KeyMan && cd ./KeyMan && git checkout -b <branch-name>
```

The subdirectories in the project root (`gatekeeper/` and `keymanager/`) house the `main.go` files for each program, however the actual functionality (and testing files) for each program is housed in the subdirectories in each of those subdirectories (`gatekeeper/gatekeeping/` and `keymanager/keymanaging/`). Use that explanation to decide what files you want to edit.

## Testing

Testing all of the programs:

```bash
rj@desktop:~/$ go test ./...
?   	github.com/the-rileyj/KeyMan/gatekeeper	[no test files]
ok  	github.com/the-rileyj/KeyMan/gatekeeper/gatekeeping	0.804s
?   	github.com/the-rileyj/KeyMan/keymanager	[no test files]
ok  	github.com/the-rileyj/KeyMan/keymanager/keymanaging	0.390s
ok  	github.com/the-rileyj/KeyMan/keymanager/utilities	0.184s
Success: Tests passed.
```

Testing the Gatekeeper program:

```bash
rj@desktop:~/gatekeeper$ go test ./...
?   	github.com/the-rileyj/KeyMan/gatekeeper	[no test files]
ok  	github.com/the-rileyj/KeyMan/gatekeeper/gatekeeping	0.804s
Success: Tests passed.
```

Testing the Keymanager program:

```bash
rj@desktop:~/keymanager$ go test ./...
ok  	github.com/the-rileyj/KeyMan/keymanager/keymanaging	0.390s
ok  	github.com/the-rileyj/KeyMan/keymanager/utilities	0.184s
Success: Tests passed.
```

Soon a `docker-compose.yml` file will be created for testing everything, for the time being this is sufficient however.

## TODO

- Docker-Compose file for testing

- Add API to gatekeeping for adding verified sources

  - Tests for this functionality
