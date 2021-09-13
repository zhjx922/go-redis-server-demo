package lib

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
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
	Data []byte
	Err error
}

type RedisMap struct {
	sync.Map
}

// Redis key和expires
type Redis struct {
	Keys RedisMap
	Expires RedisMap
	Lock sync.Mutex
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Ticker()  {
	time := time.NewTicker(1 * time.Minute)

	select {
	case <-time.C:
		r.KeysClear()
	}
}

func (r *Redis) KeysClear() {
	// 这么玩肯定会影响性能
	r.Expires.Range(func(key, expire interface{}) bool {
		if expire.(int64) < time.Now().Unix() {
			// 数据已过期
			r.Keys.Delete(key)
			r.Expires.Delete(key)
			fmt.Println("数据过期，删除")
		}
		return true
	})
}

func (r *Redis) DecodeCMD(reader io.Reader) <-chan *CMD  {
	ch := make(chan *CMD)
	go r.ParseDecode(reader, ch)
	return ch
}

func (r *Redis) Set(args []string) []byte  {
	if len(args) < 2 {
		return ReplyError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]
	l := len(args)

	fmt.Println("debug:", args)

	ex := false
	nx := false

	if l > 2 {
		for i := 2; i < l; i++ {
			a := strings.ToUpper(args[i])
			if a == "EX" {
				// 过期时间处理
				// fmt.Println("ttl:", args[i+1])
				expire, _ := strconv.ParseInt(args[i+1], 10, 64)
				r.Expires.Store(key, time.Now().Unix() + expire)
				ex = true
				i++
			} else if a == "NX" {
				nx = true
				fmt.Println("NX")
			} else {
				fmt.Println("其它命令暂不支持~")
			}
		}
	}

	if nx {
		// 加一个全局的锁
		r.Lock.Lock()
		defer r.Lock.Unlock()

		//存在直接返回nil
		if _, ok := r.GetAction(key); ok {
			fmt.Println("存在")
			return ReplyEmptyString()
		}

	}

	d := ObjectString{Data: value}
	ob := Object{Type: ObjectTypeString, Encoding: EncodingString, Data: d}
	r.Keys.Store(key, ob)

	if !ex {
		r.Expires.Delete(key)
	}

	return ReplyString("ok")
}

func (r *Redis) GetAction(key string) (string, bool)  {
	expire, ok := r.Expires.Load(key)

	if ok {
		if expire.(int64) < time.Now().Unix() {
			// 数据已过期
			r.Keys.Delete(key)
			r.Expires.Delete(key)
			return "", false
		}
	}

	if v, ok := r.Keys.Load(key); ok {
		return v.(Object).Data.(ObjectString).Data, true
	}

	return "", false
}

func (r *Redis) Get(key string) []byte  {

	if v, ok := r.GetAction(key); ok {
		return ReplyString(v)
	}

	return ReplyEmptyString()
}
