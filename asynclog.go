package main

import (
	"fmt"
	"os"
)

func NewAsyncLogWriter(fh *os.File, backlog int) *AsyncLogWriter {
	a := &AsyncLogWriter{
		F:   fh,
		buf: make(chan []byte, backlog),
	}
	go a.asyncWriter()
	return a
}

type AsyncLogWriter struct {
	F *os.File

	buf chan []byte
}

func (a *AsyncLogWriter) Write(p []byte) (n int, err error) {
	select {
	case a.buf <- p:
		return len(p), nil
	default:
		return 0, fmt.Errorf("Async log buffer full!")
	}
}

func (a *AsyncLogWriter) asyncWriter() {
	for {
		b, ok := <-a.buf
		if !ok {
			return
		}
		a.F.Write(b)
	}
}
