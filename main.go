package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("ls", "-la")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))

}
