package main

import (
	"os/exec"
	"bufio"
	"os"
	"io"
	"log"
)

type cardReader struct {
	command *exec.Cmd
	callback readerCallback
	stdoutPipe io.ReadCloser
}

type readerCallback func(string)

func startReader(cb readerCallback) (cardReader, error) {
	var (
		reader cardReader
		err error
	)

	reader.callback = cb
	reader.command = exec.Command("card-reader")
	reader.stdoutPipe, err = reader.command.StdoutPipe()

	if err == nil {
		err = reader.command.Start()
	}

	if err == nil {
		go func() {
			in := bufio.NewScanner(reader.stdoutPipe)

			for in.Scan() {
				reader.callback(in.Text())
			}

			if err := in.Err(); err != nil {
				log.Println(err)
			} else {
				log.Println("Reached EOF for reader pipe.")
			}
		}()
	}

	return reader, err
}

func (reader cardReader) stop() error {
	return reader.command.Process.Signal(os.Interrupt)
}

