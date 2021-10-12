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

type CMD struct {
	Data []byte
	Err  error
}

type RedisMap struct {
	sync.Map
}

// Redis key和expires
type Redis struct {
	Keys    map[string]interface{}
	Expires map[string]interface{}
	Lock    sync.RWMutex
}

func NewRedis() *Redis {
	return &Redis{Keys: make(map[string]interface{}), Expires: make(map[string]interface{})}
}

func (r *Redis) Ticker() {
	time := time.NewTicker(1 * time.Minute)

	select {
	case <-time.C:
		r.KeysClear()
	}
}

func (r *Redis) KeysClear() {
	// 这么玩肯定会影响性能
	for key, expire := range r.Expires {
		if expire.(int64) < time.Now().Unix() {
			// 数据已过期
			delete(r.Keys, key)
			delete(r.Expires, key)
			//r.Keys.Delete(key)
			//r.Expires.Delete(key)
			fmt.Println("数据过期，删除")
		}
	}
	/*
		r.Expires.Range(func(key, expire interface{}) bool {
			if expire.(int64) < time.Now().Unix() {
				// 数据已过期
				r.Keys.Delete(key)
				r.Expires.Delete(key)
				fmt.Println("数据过期，删除")
			}
			return true
		})
	*/
}

func (r *Redis) DecodeCMD(reader io.Reader) <-chan *CMD {
	ch := make(chan *CMD)
	go r.ParseDecode(reader, ch)
	return ch
}

func (r *Redis) SetAction(key, value string) bool {
	d := ObjectString{Data: value}
	ob := Object{Type: ObjectTypeString, Encoding: EncodingString, Data: d}
	r.Keys[key] = ob

	return true
}

func (r *Redis) Set(args []string) []byte {
	if len(args) < 2 {
		return ReplyError("ERR wrong number of arguments for 'set' command")
	}

	// 加一个全局的锁
	r.Lock.Lock()
	defer r.Lock.Unlock()

	key := args[0]
	value := args[1]
	l := len(args)

	fmt.Println("debug:", args)

	ex := false
	nx := false
	var expire int64 = 0

	if l > 2 {
		for i := 2; i < l; i++ {
			a := strings.ToUpper(args[i])
			if a == "EX" {
				// 过期时间处理
				// fmt.Println("ttl:", args[i+1])
				expire, _ = strconv.ParseInt(args[i+1], 10, 64)
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

		//存在直接返回nil
		if _, ok := r.GetAction(key); ok {
			fmt.Println("存在")
			return ReplyEmptyString()
		}

	}

	if expire > 0 {
		r.Expires[key] = time.Now().Unix() + expire
	}

	r.SetAction(key, value)

	if !ex {
		delete(r.Expires, key)
	}

	return ReplyString("OK")
}

func (r *Redis) GetAction(key string) (string, bool) {

	if expire, ok := r.Expires[key]; ok {
		if expire.(int64) < time.Now().Unix() {
			// 数据已过期
			delete(r.Keys, key)
			delete(r.Expires, key)
			return "", false
		}
	}

	if v, ok := r.Keys[key]; ok {
		return v.(Object).Data.(ObjectString).Data, true
	}

	return "", false
}

func (r *Redis) Get(key string) []byte {
	r.Lock.RLock()
	defer r.Lock.RUnlock()

	if v, ok := r.GetAction(key); ok {
		return ReplyString(v)
	}

	return ReplyEmptyString()
}

func (r *Redis) MGet(args []string) []byte {
	aLength := len(args)
	if aLength < 1 {
		return ReplyError("ERR wrong number of arguments for 'mget' command")
	}

	r.Lock.RLock()
	defer r.Lock.RUnlock()

	a := make([][]byte, aLength)

	for k, vKey := range args {
		if v, ok := r.GetAction(vKey); ok {
			a[k] = ReplyString(v)
		} else {
			a[k] = ReplyEmptyString()
		}

	}

	return ReplyArrays(a)
}

func (r *Redis) Exists(args []string) []byte {
	r.Lock.RLock()
	defer r.Lock.RUnlock()

	//返回删除成功的数量
	count := 0
	for _, vKey := range args {
		// 判断key是否存在
		if _, ok := r.Keys[vKey]; ok {
			count++
		}
	}

	return ReplyIntegers(strconv.Itoa(count))
}

func (r *Redis) Delete(args []string) []byte {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	//返回删除成功的数量
	count := 0

	for _, vKey := range args {
		// 判断key是否存在
		if _, ok := r.Keys[vKey]; ok {
			delete(r.Keys, vKey)
			delete(r.Expires, vKey)
			count++
		}
	}

	return ReplyIntegers(strconv.Itoa(count))
}

func (r *Redis) Incr(key string) []byte {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	//存在+1
	if v, ok := r.GetAction(key); ok {
		fmt.Println("存在")

		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return ReplyError("ERR 值好像不是整形哦~")
		}

		val++
		valString := strconv.FormatInt(val, 10)
		r.SetAction(key, valString)

		return ReplyIntegers(valString)
	}

	r.SetAction(key, "1")

	return ReplyIntegers("1")
}

func (r *Redis) Decr(key string) []byte {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	//存在-1
	if v, ok := r.GetAction(key); ok {
		fmt.Println("存在")

		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return ReplyError("ERR 值好像不是整形哦~")
		}

		val--
		valString := strconv.FormatInt(val, 10)
		r.SetAction(key, valString)

		return ReplyIntegers(valString)
	}

	r.SetAction(key, "-1")

	return ReplyIntegers("-1")
}
