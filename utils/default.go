package utils

import (
	"fmt"
	"reflect"
)

func PrintArr(arr any) {
	// 判断是否为slcie数据
	val := reflect.ValueOf(arr)
	if val.Kind() != reflect.Slice {
		fmt.Println("arr 不是一个切片")
		return
	}

	for i := 0; i < val.Len(); i++ {
		fmt.Println(val.Index(i).Interface())
	}

}

func StrSliceDel(s []string, idx int) []string {
	if idx < 0 || idx >= len(s) {
		return s
	}
	return append(s[:idx], s[idx+1:]...)
}
