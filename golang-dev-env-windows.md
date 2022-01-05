# Installing a Go development environment on a Windows PC

*NOTE*: A lot of these instructions change with Go 1.11 and later. These instructions are for Go 1.10.

## Prerequisites

Install the following programs.

* Git bash (https://git-scm.com/download/win)
  Git is necessary for the `go get` command, and using the git bash command line works well with the go dev environment.
  
* ~~MSYS2 (https://www.msys2.org/)
  This is necessary if you want to install packages that include cgo (Go linked with C). SQLite is one common package
  that requires a C compiler to build. *NOTE:* Be sure to install the 64 bit version. By default it installs to `C:\msys64`.~~
  
* TDM-GCC (http://tdm-gcc.tdragon.net/download)
  This is necessary if you want to install packages that include cgo (Go linked with C). SQLite is one common package
  that requires a C compiler to build. *NOTE:* Be sure to install the 64 bit version. By default it installs to `C:\TDM-GCC-64`.
  
  
* GNU Make (http://gnuwin32.sourceforge.net/packages/make.htm)
  This is just handy. It is not uncommon to create a simple Makefile for a project.
  
  Remember to add the following directory to your PATH
  `C:\Program Files (x86)\GnuWin32\bin`

## Install Go SDK

Download and install MSI from https://golang.org/dl

## Create GOPATH environment variable

Control Panel | System And Security | System | Advanced System Settings | Environment Variables ...

Set the following environment variables for the user:

```
GOPATH=D:\go
```

## Create GOPATH directories

Create a directory for the GOPATH, these instructions assume it is at `D:\go`

1. Create GOPATH directory

```bash
mkdir -p $GOPATH/src $GOPATH/pkg $GOPATH/bin
```

## Add GOPATH/bin to your path

Control Panel | System And Security | System | Advanced System Settings | Environment Variables ...

Add the following directory to the user PATH:
```
D:\go\bin
```

Any commands that you compile using the `go install` command will be saved to the `$GOPATH/bin` directory. Putting this
directory on the path makes it easy to run commands.

*NOTE:* Take care installing unknown packages -- they could install a command. Make sure that `$GOPATH/bin`
is at the end of the path, not the beginning. That way nothing in `$GOPATH/bin` will take precedence over
system commands.

## Add dep command

Download the latest `dep` executable for 64 bit windows from https://github.com/golang/dep/releases and
install in a directory on the PATH. (I use `$GOPATH/bin`, but anywhere on the path will do).

*NOTE:* The `dep` command is used for package management. It was part of an 'official experiment', but 
is replaced in Go 1.11 with a completely different package management solution. This is the subject of
a reasonable amount of controversy in the Gopher world. The `dep` command is deprecated as of Go 1.11, 
but until that is out we are still making use of it for package management.

## Add sundry useful commands

Install some useful command line tools.

```bash
go get -u golang.org/x/tools/cmd/godoc
```

The Go SDK supports code generation "to automate the running of tools to generate source code before compilation".
Support for code generation is via the `go generate` command.

There are lots of little tools that are handy for code generation (and it is possible to write your own).
It is a good idea to install the following commonly used code generation tools:

```bash
go get -u golang.org/x/tools/cmd/stringer
go get -u github.com/alvaroloes/enumer
```

## Install VS Code

There are a number of good options for a Go IDE, but VS Code is probably the pick of the bunch at the moment.
There are, of course, different opinions. Feel free to choose another (eg JetBrains GoLand).

To install VS Code download from https://code.visualstudio.com/download.

Once installed, there are a number of extensions to install. To install an extension:

* Press Ctrl-Shift-P
* Type "Install Extensions" and press enter. This will open the extensions sidebar. Search through the extensions for the one you want and press "Install".

Extensions to install are:

| Name        | Author    | Extension Identifer | Comments |
| ------------|-----------|---------------------|----------|
| Go          | Microsoft | ms-vscode.go        | This is the Go extension. Necessary for Go development. |
| Better TOML | bungcip   | bungcip.better-toml | Handy for editing TOML files. |