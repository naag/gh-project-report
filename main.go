package main

import "github.com/naag/gh-project-report/cmd"

var (
	Version   string
	Commit    string
	BuildTime string
)

func main() {
	cmd.Execute()
}
