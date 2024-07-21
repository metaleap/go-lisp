package main

import (
	"io"
)

func readUntil(r io.Reader, until byte, initialBufCapacity int) (string, error) {
	buf := make([]byte, 0, initialBufCapacity)
	var b [1]byte
	for {
		_, err := r.Read(b[0:1])
		if err != nil {
			return "", err
		} else if b[0] == until {
			break
		} else {
			buf = append(buf, b[0])
		}
	}
	line := string(buf)
	return line, nil
}
