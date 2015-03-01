package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

func logger(wg *sync.WaitGroup) chan<- string {
	wg.Add(1)
	var out *bufio.Writer

	logfile, err := os.Create("log.txt")
	if err != nil {
		out = bufio.NewWriter(os.Stdout)
	} else {
		out = bufio.NewWriter(logfile)
	}

	logchan := make(chan string)

	go func() {
		defer wg.Done()
		out.WriteString(time.Now().String() + "\n")
		for {
			msg, ok := <-logchan
			if !ok {
				out.Flush()
				logfile.Close()
				return
			}
			out.WriteString(fmt.Sprintf("%s\n", msg))
		}
	}()

	return logchan
}
