package common_log

import (
	"testing"
)

// 测试用结构体
type Data struct {
	Name   string `trans:"姓名" log:"need"`
	Age    int    `log:"need"`
	Class  string `trans:"班级" log:"special"`
	Sex    int    `trans:"性别"  log:"need"`
	Weight int    `trans:"体重"`
}

// 用于特殊结构体字段转换，命名规则为Change+fieldName
func (d *Data) ChangeSex() string {
	var sexMap = map[int]string{1: "男", 2: "女"}
	return sexMap[d.Sex]
}

func Test_DiffAndMakeInfo(t *testing.T) {

	oldData := Data{
		Name:   "大狗子",
		Age:    0,
		Class:  "五班",
		Sex:    2,
		Weight: 45,
	}

	newData := Data{
		Name:   "大狗子",
		Age:    2,
		Class:  "三班",
		Sex:    1,
		Weight: 50,
	}

	info, err := NewCommonLog().MakeInfo(&oldData, &newData,
		Extra{
			Name:    "测试Extra的新值",
			OldData: nil,
			NewData: "新值",
		},
		Extra{
			Name:    "测试Extra的旧值",
			OldData: "旧值",
			NewData: nil,
		},
		Extra{
			Name:    "测试Extra的变化",
			OldData: "A",
			NewData: "B",
		},
	)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("（数据变更）", info)

	createData := Data{
		Name:   "大狗子",
		Age:    0,
		Class:  "五班",
		Sex:    2,
		Weight: 45,
	}
	createInfo, err := NewCommonLog().MakeInfo(nil, &createData)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("（数据新增）", createInfo)

	delData := Data{
		Name:   "大狗子",
		Age:    2,
		Class:  "三班",
		Sex:    1,
		Weight: 50,
	}
	delInfo, err := NewCommonLog().MakeInfo(&delData, nil)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("（数据删除）", delInfo)
}
