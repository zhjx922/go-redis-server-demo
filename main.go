package main

import (
	"fmt"
	"go-redis-server-demo/lib"
	"log"
	"net"
)

func main()  {
	fmt.Println("Redis Server Start……")

	// 启动TCP Server 监听6379端口
	listener, err := net.Listen("tcp", "127.0.0.1:6379")

	if err != nil {
		log.Println(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handler(conn)
	}
}

// handler 处理数据
func handler(conn net.Conn)  {
	defer conn.Close()
	for {
		cmd := lib.DecodeCMD(conn)

		for c := range cmd {
			if c.Err == nil {
				conn.Write([]byte("+OK\r\n"))
			} else {
				fmt.Println("异常了")
			}
		}
	}
}