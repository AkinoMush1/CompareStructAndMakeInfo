package common_log

type Extra struct {
	Name    string      // 字段别名
	OldData interface{} // 老值
	NewData interface{} // 新值
}
