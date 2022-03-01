package common_log

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type CommonLog struct {
}

func NewCommonLog() *CommonLog {
	return &CommonLog{}
}

// MakeInfo 生成日志
// oldStruct和newStruct需要为指针类型
// 其他使用说明请看README文件
func (c *CommonLog) MakeInfo(oldStruct, newStruct interface{}, extras ...Extra) (info string, err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			info = "生成日志信息从崩溃中恢复，不记录差异"
			err = errors.New("生成日志信息从崩溃中恢复，不记录差异")
		}
	}()

	info1, err := c.compareStruct(oldStruct, newStruct)
	if err != nil {
		return "", err
	}
	info2 := c.makeExtraInfo(extras)
	if info1 != "" && info2 != "" {
		info = info1 + "; " + info2
	} else if info1 != "" {
		info = info1
	} else if info2 != "" {
		info = info2
	} else {
		info = "无任何关键数据发生变更"
	}
	return
}

// compareStruct 比较新旧数据的差异，并返回差异字段信息
func (c *CommonLog) compareStruct(oldData, newData interface{}) (info string, err error) {
	if oldData != nil && reflect.TypeOf(oldData).Kind() != reflect.Ptr {
		return "", errors.New("传入的结构体必须为指针类型")
	}

	if newData != nil && reflect.TypeOf(newData).Kind() != reflect.Ptr {
		return "", errors.New("传入的结构体必须为指针类型")
	}

	if oldData == nil && newData == nil {
		return "", nil
	} else if oldData == nil {
		return c.reflectStruct(newData), nil
	} else if newData == nil {
		return c.reflectStruct(oldData), nil
	}

	return c.diffStruct(oldData, newData), nil
}

// diffStruct 比较两个结构体数据
func (c *CommonLog) diffStruct(oldData, newData interface{}) (info string) {
	var (
		oldRefElem = reflect.ValueOf(oldData).Elem()
		newRefElem = reflect.ValueOf(newData).Elem()
		refType    = newRefElem.Type()
		msgs       = make([]string, 0)
	)

	for i := 0; i < newRefElem.NumField(); i++ {
		n := refType.Field(i).Name
		t := refType.Field(i).Tag

		// 根据model的logTag标识来判断是否需要进行比较
		if t.Get("log") == "" {
			continue
		}

		trans := t.Get("trans")
		if trans == "" {
			trans = n
		}

		oldValue := oldRefElem.FieldByName(n).Interface()
		newValue := newRefElem.FieldByName(n).Interface()
		if !reflect.DeepEqual(oldValue, newValue) {
			msg := ""
			if t.Get("log") == "special" {
				msg = fmt.Sprintf("%v 发生了变更", trans)
			} else {
				msg = c.makeMsgWithCheckingBlank(reflect.ValueOf(oldData), reflect.ValueOf(newData),
					oldRefElem.FieldByName(n), newRefElem.FieldByName(n), n, trans)
			}
			if msg == "" {
				continue
			}
			msgs = append(msgs, msg)
		}
	}

	return strings.Join(msgs, "; ")
}

// reflectStruct 返回结构体数据
func (c *CommonLog) reflectStruct(data interface{}) (info string) {
	var (
		newRefElem = reflect.ValueOf(data).Elem()
		refType    = newRefElem.Type()
		msgs       = make([]string, 0)
	)

	for i := 0; i < newRefElem.NumField(); i++ {
		n := refType.Field(i).Name
		t := refType.Field(i).Tag

		if t.Get("log") == "" {
			continue
		}

		trans := t.Get("trans")
		if trans == "" {
			trans = n
		}

		msg := ""
		if t.Get("log") == "special" {
			msg = fmt.Sprintf("%v 发生了变更", trans)
		} else {
			// 因为reflect是通用方法，所以这里传入newStruct或者oldStruct都行，
			msg = c.makeMsgWithCheckingBlank(reflect.ValueOf(nil), reflect.ValueOf(data),
				reflect.ValueOf(nil), newRefElem.FieldByName(n), n, trans)
		}
		if msg == "" {
			continue
		}
		msgs = append(msgs, msg)
	}

	return strings.Join(msgs, "; ")
}

// makeExtraInfo 追加额外信息
func (c *CommonLog) makeExtraInfo(extras []Extra) (info string) {
	msgs := make([]string, 0)
	for _, extra := range extras {
		if extra.OldData != nil && extra.NewData != nil {
			if !reflect.DeepEqual(extra.OldData, extra.NewData) {
				msg := c.makeMsgWithCheckingBlank(reflect.ValueOf(nil), reflect.ValueOf(nil),
					reflect.ValueOf(extra.OldData), reflect.ValueOf(extra.NewData), "extra", extra.Name)
				if msg == "" {
					continue
				}
				msgs = append(msgs, msg)
			}
		} else {
			msg := c.makeMsgWithCheckingBlank(reflect.ValueOf(nil), reflect.ValueOf(nil),
				reflect.ValueOf(extra.OldData), reflect.ValueOf(extra.NewData), "extra", extra.Name)
			if msg == "" {
				continue
			}
			msgs = append(msgs, msg)
		}
	}

	return strings.Join(msgs, "; ")
}

// makeMsgWithCheckingBlank 判断空值并生成对应的msg
func (c *CommonLog) makeMsgWithCheckingBlank(oldStructV, newStructV, oldValue, newValue reflect.Value,
	fieldName, name string) (msg string) {
	if oldValue.IsValid() && newValue.IsValid() {
		oldV := changeValue(oldStructV, oldValue, fieldName)
		newV := changeValue(newStructV, newValue, fieldName)

		// 补偿措施，映射后的值相同，则也记为无变更
		v1, ok1 := oldV.(reflect.Value)
		v2, ok2 := newV.(reflect.Value)
		if ok1 && ok2 && reflect.DeepEqual(v1, v2) {
			return ""
		}

		msg = fmt.Sprintf("%v 从 %v 变更为 %v", name, oldV, newV)
	} else if oldValue.IsValid() {
		v := changeValue(oldStructV, oldValue, fieldName)
		msg = fmt.Sprintf("%v 值为 %v ", name, v)
	} else if newValue.IsValid() {
		v := changeValue(newStructV, newValue, fieldName)
		msg = fmt.Sprintf("%v 值为 %v ", name, v)
	}

	return
}

// changeValue 根据函数或者空值更改值
func changeValue(structV, value reflect.Value, fieldName string) interface{} {
	v := value.Interface()

	if value.IsZero() {
		v = reflect.ValueOf("空")
	} else {
		if structV.IsValid() {
			f1 := structV.MethodByName("Change" + fieldName)
			if f1.IsValid() {
				v = f1.Call([]reflect.Value{})[0]
				if v1, ok := v.(reflect.Value); ok {
					if v1.IsZero() {
						v = reflect.ValueOf("空")
					}
				}
			}
		}
	}

	return v
}
