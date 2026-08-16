package main

import "github.com/ErKatta/zbxctl/cmd"

var (
	version = "0.2.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if version != "" && version != "0.2.0-dev" {
		cmd.Version = version
		cmd.RootCmd.Version = version
	}
	if commit != "" && commit != "none" {
		cmd.Commit = commit
	}
	if date != "" && date != "unknown" {
		cmd.BuildDate = date
	}
	cmd.Execute()
}
