package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintf(os.Stdout, "Hello on Stdout\n")
	fmt.Fprintf(os.Stderr, "Hello on Stderr\n")
}
