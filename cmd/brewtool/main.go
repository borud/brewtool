package main

import (
	"github.com/borud/brewtool/pkg/util"
)

var opt struct {
	Owner string `long:"owner" default:"borud" description:"repository owner" required:"yes"`
	Repo  string `long:"repo" description:"repository" required:"yes"`

	Generate generateCmd `command:"generate" alias:"gen" description:"generate brew file"`
}

func main() {
	util.FlagParse(&opt)
}
