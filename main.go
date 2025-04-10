package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// cmd.Execute()
	cur, err := os.Getwd()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Print("===curl>:", cur)
	dir, file := filepath.Split(cur)
	fmt.Printf("===> dir:%s, file: %s\n", dir, file)
	names := strings.Split(cur, string(os.PathSeparator))
	fmt.Printf("===> names: %#v, len: %d\n", names, len(names))
}
