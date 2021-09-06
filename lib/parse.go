package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
)

type Command struct {
	TmpBuffer []byte
	MsgType byte //*,$,+,-,:
	Len int //数组长度or字符串长度
	Read bool //是否继续read
	Argv []string //参数数据
}

// ParseDecode 协议解析
// 参考文章 https://www.redis.com.cn/topics/protocol.html
func (r *Redis) ParseDecode(reader io.Reader, ch chan<- *CMD){
	read := bufio.NewReader(reader)

	command := &Command{MsgType: 0, Len: 0, Read: false}

	for {
		// 貌似有点浪费空间
		buf := make([]byte, 1024)
		length := 0

		//buffer没有内容了
		if command.Read {
			var err error
			length, err = read.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}
		}

		r.ParseCommand(buf[:length], command, ch)
	}
}

// ParseCommand 解析一条命令
func (r *Redis) ParseCommand(b []byte, c *Command, ch chan<- *CMD)  {
	c.TmpBuffer = append(c.TmpBuffer, b...)

	index := bytes.IndexByte(c.TmpBuffer, '\n')
	if index == -1 {
		//数据还不够,继续读
		c.Read = true
		return
	}

	if c.MsgType == 0 {
		//未解析，需要解析首行
		r.ParseCommandFirst(c, ch)
	} else if c.MsgType == TypeArray {
		//遍历数据，获取
		r.ParseCommandArray(c, ch)
	} else if c.MsgType == TypeStringBulk {
		//多行数据获取
	}
}

// ParseCommandFirst 首行命令解析
func (r *Redis) ParseCommandFirst(c *Command, ch chan<- *CMD)  {
	index := bytes.IndexByte(c.TmpBuffer, '\n')

	// 判断命令类型
	c.MsgType = c.TmpBuffer[0]
	value := c.TmpBuffer[1:index-1]
	fmt.Println("FirstMsgValue:", string(value))

	if c.MsgType == TypeArray {
		i ,err := strconv.Atoi(string(value))
		if err == nil {
			fmt.Println("数组长度:", i)
			c.Len = i
		}
	}

	//缓存剩余数据
	if (index + 1) <= len(c.TmpBuffer) {
		//暂停从io读取数据
		c.Read = false
		c.TmpBuffer = c.TmpBuffer[index+1:]
	} else {
		c.Read = true
		c.TmpBuffer = []byte{}
	}
}

func (r *Redis) ParseCommandArray(c *Command, ch chan<- *CMD)  {
	read := true
	index := bytes.IndexByte(c.TmpBuffer, '\n')
	value := c.TmpBuffer[1:index-1]
	fmt.Println("array:", string(c.TmpBuffer))

	// 多行字符串处理
	if c.TmpBuffer[0] == TypeStringBulk {
		i ,err := strconv.Atoi(string(value))
		if err == nil {
			fmt.Println("字符串长度:", i)
		}
		sLength := i + index + 1 + 2

		fmt.Println(len(c.TmpBuffer), sLength)
		if len(c.TmpBuffer) >= sLength {
			fmt.Println("数据够了:", string(c.TmpBuffer[index+1:index+1+i]))

			c.Argv = append(c.Argv, string(c.TmpBuffer[index+1:index+1+i]))

			if len(c.TmpBuffer) == sLength {
				c.TmpBuffer = []byte{}
			} else {
				c.TmpBuffer = c.TmpBuffer[sLength:]
				read = false
				fmt.Println("剩余数据：", string(c.TmpBuffer))
			}

		} else {
			fmt.Println("数据不够，继续获取")
		}
	} else if c.TmpBuffer[0] == TypeString {
		c.Argv = append(c.Argv, string(c.TmpBuffer[1:index-2]))

		if len(c.TmpBuffer) == (index + 1) {
			c.TmpBuffer = []byte{}
		} else {
			c.TmpBuffer = c.TmpBuffer[index+1:]
			read = false
			fmt.Println("剩余数据：", string(c.TmpBuffer))
		}

	} else {
		fmt.Println("特殊情况？？")
	}

	if len(c.Argv) == c.Len {
		fmt.Println("开始执行命令")

		for _, v := range c.Argv {
			fmt.Println(v)
		}

		reply := "ok"
		if c.Argv[0] == "set" {
			r.Set(c.Argv[1], c.Argv[2])
		} else if c.Argv[0] == "get" {
			reply = r.Get(c.Argv[1])
		}

		c.MsgType = 0
		c.Len = 0
		c.TmpBuffer = []byte{}
		c.Argv = []string{}
		read = true

		//Reply
		ch <- &CMD{
			Data: reply,
			Err: nil,
		}
	}

	c.Read = read
}

func ParseString() {

}
