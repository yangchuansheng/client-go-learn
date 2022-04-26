package util

import (
	"bufio"
	"log"
	"os"
)

func Prompt() {
	log.Printf("-> Press Enter to continue...")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	log.Println()
}
