package utils

import (
	"encoding/json"
	"fmt"
	"medialpha-backend/constant"
	"reflect"
	"strings"
)

// Wrapper
// @Description   封装接口返回数据
// @Author        xiaolong
// @Date          2022/11/29 16:11(create);
// @Param         message   string   返回消息
// @Param         code      int      返回代码
// @Param         data     	any      携带数据
// @Return        json []byte
func Wrapper(message string, code int, data any) []byte {
	m := map[string]any{
		"message": message,
		"code":    code,
		"data":    data,
	}
	bytes, _ := json.Marshal(m)
	return bytes
}

func WrapperOK(data any) []byte {
	return Wrapper("ok", constant.Resp.OK(), data)
}

func WrapperError(message string) []byte {
	return Wrapper(message, constant.Resp.Fail(), struct{}{})
}

func WrapperErrorNoAuth() []byte {
	return Wrapper("无权限访问", constant.Resp.Fail(), struct{}{})
}

func WrapperParamIllegal() []byte {
	return Wrapper("参数非法", constant.Resp.Fail(), struct{}{})
}

func ID2Str(id any) string {
	return fmt.Sprintf("%d", id)
}

type TypeFilter func(key string, val any) any

func ID2StrFilter(key string, val any) any {
	if strings.HasSuffix(strings.ToLower(key), "id") {
		return fmt.Sprintf("%d", val)
	}
	return val
}

func Struct2Map(o any, filters ...TypeFilter) map[string]any {
	// 通过反射将结构体转换成map
	data := make(map[string]any)
	objT := reflect.TypeOf(o)
	objV := reflect.ValueOf(o)
	if objT.Kind() == reflect.Pointer {
		objV = objV.Elem()
		objT = objT.Elem()
	}
	for i := 0; i < objT.NumField(); i++ {
		jsonTag, ok := objT.Field(i).Tag.Lookup("json")
		if jsonTag == "-" {
			continue
		}
		if ok {
			val := objV.Field(i).Interface()
			for _, filter := range filters {
				val = filter(jsonTag, val)
			}
			data[jsonTag] = val
		} else {
			name := objT.Field(i).Name
			val := objV.Field(i).Interface()
			for _, filter := range filters {
				val = filter(name, val)
			}
			data[name] = val
		}
	}
	return data
}

func Structs2Maps(o any, filters ...TypeFilter) []*map[string]any {
	// 通过反射将结构体转换成map
	var res []*map[string]any

	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Slice {
		return make([]*map[string]any, 0)
	}
	if v.Len() <= 0 {
		return make([]*map[string]any, 0)
	}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i).Interface()
		m := Struct2Map(elem, filters...)
		res = append(res, &m)
	}
	return res
}

func SafeSlice(s any) any {
	if reflect.TypeOf(s).Kind() != reflect.Slice {
		return s
	}
	if reflect.ValueOf(s).Len() <= 0 {
		return make([]struct{}, 0)
	}
	return s
}
