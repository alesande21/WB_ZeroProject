package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := os.Stat("file.txt")
	stat := os.IsNotExist(err)
	fmt.Println(!stat)
}
