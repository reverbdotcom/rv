RV
---

`rv` is a quick AWS EC2 instance resolver

## Usage
```BASH
NAME:
   rv - node resolver

USAGE:
   rv [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   ip, i
   list, l
   cmd, c
   help, h  Shows a list of commands or help for one command
```

### ip
Gathers the private ip of the given named instance.

### cmd
Runs the given command with the named instances replaced for the ips.

### list
Lists all the current instances.

## Install
`brew install ebenoist/rv`
