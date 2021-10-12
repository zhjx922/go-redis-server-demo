package lib

const (
	// ObjectTypeString 字符串
	ObjectTypeString = 1
	// ObjectTypeList 列表
	ObjectTypeList = 2
	// ObjectTypeHash 哈希
	ObjectTypeHash = 3
	// ObjectTypeSet 集合
	ObjectTypeSet = 4
	// ObjectTypeZSet 有序集合
	ObjectTypeZSet = 5
)

const (
	EncodingString = 1
	EncodingInt    = 2
)

// Object 单个数据对象
type Object struct {
	// 类型
	Type int
	// 编码类型
	Encoding int
	// 数据
	Data interface{}
}

type ObjectString struct {
	Data string
}

type ObjectInt struct {
	Data int
}
