package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello world!")
	fn := filepath.Join(os.Getenv("HOME"), "foo.txt")
	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	fmt.Printf("Opened %s\n", fn)

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil {
		fmt.Printf("read: %v\n", err)
	} else {
		fmt.Println(string(buf[:n]))
	}
}
