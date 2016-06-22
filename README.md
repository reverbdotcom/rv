RV
---

`rv` is a quick AWS EC2 instance resolver. Take your cloud with you wherever you are.

![rv](http://i.imgur.com/XHNglPk.jpg?1)

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
* `brew tap reverbdotcom/rv`
* `brew install rv`

You will also need to tell the AWS client which region to default to, you can do this by exporting AWS_REGION.

For bash:
```
$ echo 'export AWS_REGION=us-east-1' >> ~/.bashrc
```

For zsh:
```
$ echo 'export AWS_REGION=us-east-1' >> ~/.zshrc
```

## Release

1. Create a new github release with a version that starts with 'v' such as 'v0.0.7'
2. The travis build for this repository will automatically tar up a binaries.tar.gz and attach it to a github release.

If the travis build is not working correctly, you can run the same binary tar creation script from the .travis.yml (before_deploy) and upload the binaries to the release manually on GitHub.

## LICENSE

Copyright 2016 Reverb.com, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
