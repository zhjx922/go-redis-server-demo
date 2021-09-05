package lib

import (
	"io"
)

// 命令类型
const (
	TypeString byte = '+'
	TypeError byte = '-'
	TypeInt byte = ':'
	TypeStringBulk byte = '$'
	TypeArray byte = '*'
)

type CMD struct{
	Data string
	Err error
}

type Redis struct {

}

func DecodeCMD(reader io.Reader) <-chan *CMD  {
	ch := make(chan *CMD)
	go ParseDecode(reader, ch)
	return ch
}

func (r *Redis) Set(key, value string)  {

}
