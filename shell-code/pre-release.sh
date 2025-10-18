#!/bin/bash

go-ansible-vault -i shell-code/build.env -a get -m GITHUB_TOKEN  > /tmp/source && source /tmp/source

GITHUB_TOKEN=$GITHUB_TOKEN python3 shell-code/release.py