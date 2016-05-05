package testdata

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func checkingScannerErr() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		fmt.Println(s.Text())
	}
	if err := s.Err(); err != nil {
		log.Println(err)
	}
}

func notCheckingScannerErr() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		fmt.Println(s.Text())
	}
	fmt.Println("YOLO")
}
