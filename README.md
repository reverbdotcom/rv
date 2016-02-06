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
   0.0.3

COMMANDS:
   ip, i
   list, l
   cmd, c
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --clear-cache, -c	ensure rv cache is cleared
   --help, -h		show help
   --version, -v	print the version
```

### ip
Gathers the private ip of the given named instance.
```Bash
$ rv i mynode.local
127.0.0.1
```

### cmd
Runs the given command with the named instances replaced for the ips.
```Bash
$ rv c "ssh user@my-node.local"
Welcome to Ubuntu 14.04.1 LTS (GNU/Linux 3.13.0-43-generic x86_64)
```

### list
Lists all the current instances.
```
$ rv l

Name                             Private IP Address
mynode.local                     127.0.0.1
mynode-2.local                   127.0.0.1
```


### cache
rv will cache responses from AWS for a minute. Running the command with the --clear-cache flag will ensure a cache miss.

## Install
* `brew tap ebenoist/rv`
* `brew install rv`
