#!/bin/bash

target=$1
fast=$2

bin=notify4g
#bin=`basename "$PWD"`

echo bin:${bin}
echo targetos:${target}
echo fastmode:${fast}

# notice how we avoid spaces in $now to avoid quotation hell in go build
now=$(date +'%Y-%m-%d_%T')

if [[ ${target} = "linux" ]]; then
    echo "build for linux"
    env GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$now" -o "${bin}_linux_amd64"
    if [[ ${fast} != "fast" ]] && type upx > /dev/null 2>&1; then
        upx ${bin}_linux_amd64
    fi
else
    echo "build for local"
    go build -ldflags "-w -s -X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$now"

    if [[ ${fast} != "fast" ]] && type upx > /dev/null 2>&1; then
        upx rig
    fi
fi


# meaning of -ldflags '-w -s'
# https://stackoverflow.com/questions/22267189/what-does-the-w-flag-mean-when-passed-in-via-the-ldflags-option-to-the-go-comman
# You will get the smallest binaries if you compile with -ldflags '-w -s'.
# The -w turns off DWARF debugging information: you will not be able to use gdb on the binary to
# look at specific functions or set breakpoints or get stack traces, because all the metadata gdb
# needs will not be included. You will also not be able to use other tools that depend on the information,
# like pprof profiling. The -s turns off generation of the Go symbol table:
# you will not be able to use 'go tool nm' to list the symbols in the binary.
# Strip -s is like passing -s to -ldflags but it doesn't strip quite as much.
# 'Go tool nm' might still work after 'strip -s'. I am not completely sure.

# None of these - not -ldflags -w, not -ldflags -s, not strip -s - should affect the execution of the actual program.
# They only affect whether you can debug or analyze the program with other tools.

# $ go tool link
#   ...
#   -s    disable symbol table
#   -w    disable DWARF generation

