package lib

import (
	"bytes"
	"strconv"
)

func ReplyString(data string) []byte {
	return []byte("+" + data + "\r\n")
}

func ReplyEmptyString() []byte {
	return []byte("$-1\r\n")
}

func ReplyIntegers(i string) []byte {
	return []byte(":" + i + "\r\n")
}

func ReplyArrays(r [][]byte) []byte {
	l := len(r)
	lString := strconv.Itoa(l)
	var b bytes.Buffer
	a := []byte("*" + lString + "\r\n")

	b.Write(a)
	for _, v := range r {
		b.Write(v)
	}

	return b.Bytes()
}

func ReplyError(data string) []byte {
	return []byte("-" + data + "\r\n")
}
