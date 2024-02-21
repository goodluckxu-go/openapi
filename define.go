package openapi

const (
	firstKeyValueCutSign  = ":"  // 第一次切割键值标志
	secondListCutSign     = ";"  // 第二次切割数组标志
	secondKeyValueCutSign = "="  // 第二次切割键值标志
	thirdListCutSign      = ","  // 第三次切割数组标志
	multiBorderSign       = "|-" // 多行标志
	multiBorderSignEnd    = "-|" // 多行标志结束，不存在则在下一个可用标签前结束

	structDeep = 10 // 多重结构体注释递归深度，防止出现链式调用
)

const (
	validTypeString = iota
	validTypeMap
	validTypeMapArray
	validTypeArray
	validTypeBool
	validTypeInteger
	validTypeJson
)

var (
	projectModName = ""
	modPathMap     modHandle
)
