package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	r := bufio.NewReader(os.Stdin)

	// create a new scanner
	scanner := bufio.NewScanner(r)

	// Use scanword to split
	scanner.Split(bufio.ScanWords)
	words := 0
	for scanner.Scan() {
		words++
		// fmt.Println(scanner.Text())
	}
	fmt.Printf("%d\n", words)

	// check for the error that occurred during the scanning
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
