package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func GetQuery(req *http.Request, key string) (string, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return "", fmt.Errorf("[%v]该参数不存在", key)
	}
	return vals[0], nil
}

func GetQuerySlice(req *http.Request, key string) ([]string, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return nil, fmt.Errorf("[%v]该参数不存在", key)
	}
	return vals, nil
}

func GetQueryInt64(req *http.Request, key string) (int64, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return 0, fmt.Errorf("[%v]该参数不存在", key)
	}

	val, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("[%v]该参数值类型非法", key)
	}
	return val, nil
}

func GetQueryInt64Slice(req *http.Request, key string) ([]int64, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return nil, fmt.Errorf("[%v]该参数不存在", key)
	}

	var res []int64
	for _, val := range vals {
		valInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("[%v]该参数值类型非法", key)
		}
		res = append(res, valInt)
	}
	return res, nil
}

func GetQueryInt(req *http.Request, key string) (int, error) {
	res, err := GetQueryInt64(req, key)
	return int(res), err
}

func GetQueryIntSlice(req *http.Request, key string) ([]int, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return nil, fmt.Errorf("[%v]该参数不存在", key)
	}

	var res []int
	for _, val := range vals {
		valInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("[%v]该参数值类型非法", key)
		}
		res = append(res, int(valInt))
	}
	return res, nil
}

func GetQueryFloat(req *http.Request, key string) (float64, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return 0, fmt.Errorf("[%v]该参数不存在", key)
	}

	val, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		return 0, fmt.Errorf("[%v]该参数值类型非法", key)
	}
	return val, nil
}

func GetQueryFloatSlice(req *http.Request, key string) ([]float64, error) {
	q := req.URL.Query()
	vals, exists := q[key]
	if !exists || len(vals) == 0 {
		return nil, fmt.Errorf("[%v]该参数不存在", key)
	}

	var res []float64
	for _, val := range vals {
		valInt, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("[%v]该参数值类型非法", key)
		}
		res = append(res, valInt)
	}
	return res, nil
}

func ParseBody(req *http.Request, ptr any) error {
	jsonBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, ptr)
	return err
}

func GetBody(req *http.Request, key string) (string, error) {
	var bodyMap map[string]any
	err := ParseBody(req, &bodyMap)
	if err != nil {
		return "", err
	}
	val, exists := bodyMap[key]
	if !exists {
		return "", fmt.Errorf("key: %v不存在", key)
	}
	valStr, _ := val.(string)
	return valStr, nil
}
