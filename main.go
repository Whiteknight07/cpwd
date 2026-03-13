package main

import (
	"os"

	"github.com/Whiteknight07/cpwd/internal/cpwd"
)

func main() {
	os.Exit(cpwd.Run(os.Stdout, os.Stderr))
}
