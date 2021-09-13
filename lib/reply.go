package lib


func ReplyString(data string) []byte  {
	return []byte("+" + data +"\r\n")
}

func ReplyEmptyString() []byte  {
	return []byte("$-1\r\n")
}

func ReplyError(data string) []byte  {
	return []byte("-" + data +"\r\n")
}