package main

import (
	"time"

	"github.com/ml444/gctl/cmd"
)

func main() {
	cmd.Execute()
	time.Sleep(time.Millisecond * 100)
}
