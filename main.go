package main

import "github.com/ismaels/git-ssh-sign/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
