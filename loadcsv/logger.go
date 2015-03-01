package main

import (
	"bufio"
	"fmt"
	"os"
)

func logger(done <-chan bool) chan<- string {
	var out *bufio.Writer

	logfile, err := os.Create("log.txt")
	if err != nil {
		out = bufio.NewWriter(os.Stdout)
	} else {
		out = bufio.NewWriter(logfile)
	}

	logchan := make(chan string)

	go func() {
		defer logfile.Close()
		defer out.Flush()
		defer close(logchan)
	loop:
		for {
			select {
			case msg := <-logchan:
				out.WriteString(fmt.Sprintf("%s\n", msg))
			case <-done:
				break loop
			}
		}
	}()

	return logchan
}
