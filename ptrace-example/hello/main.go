package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello world!")
	fn := "/etc/passwd"
	if _, err := os.Open(fn); err != nil {
		fmt.Printf("open: %v", err)
	}
	fmt.Printf("Opened %s\n", fn)
}
