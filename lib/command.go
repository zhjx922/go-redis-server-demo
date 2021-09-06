package lib

import (
	"io"
	"sync"
)

// 命令类型
const (
	// TypeString 单行字符串
	TypeString byte = '+'
	// TypeError 错误信息
	TypeError byte = '-'
	// TypeInt 整形
	TypeInt byte = ':'
	// TypeStringBulk 多行字符串
	TypeStringBulk byte = '$'
	// TypeArray 数组
	TypeArray byte = '*'
)

type CMD struct{
	Data string
	Err error
}

// Data数据放在这里
type Data struct {
	Type byte //数据类型
	Value interface{}
}

type Redis struct {
	Keys sync.Map
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) DecodeCMD(reader io.Reader) <-chan *CMD  {
	ch := make(chan *CMD)
	go r.ParseDecode(reader, ch)
	return ch
}

func (r *Redis) Set(key, value string)  {
	r.Keys.Store(key, value)
}

func (r *Redis) Get(key string) string  {
	v, ok := r.Keys.Load(key)

	if ok {
		return v.(string)
	}

	return ""
}
