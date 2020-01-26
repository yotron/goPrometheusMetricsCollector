package jsontools

import (
	"encoding/json"
	. "github.com/yotron/goConfigurableLogger"
	"strconv"
	"strings"
)

func GetDataOfSlice(data map[string]interface{}, resultpath string, index int64) string {
	pathValue := strings.Split(resultpath, ".")
	value := data[pathValue[index]]
	Debug.Println("Index:", index, "PathValue:", pathValue, "ValueType:", value)
	switch value.(type) {
	case int:
		Debug.Println("value", value, " is int")
		return strconv.Itoa(value.(int))
	case int8:
		Debug.Println("value", value, " is int8")
		return strconv.Itoa(value.(int))
	case int16:
		Debug.Println("value", value, " is int16")
		return strconv.Itoa(value.(int))
	case int32:
		Debug.Println("value", value, " is int32")
		return strconv.Itoa(value.(int))
	case int64:
		Debug.Println("value", value, " is int64")
		return strconv.Itoa(value.(int))
	case string:
		Debug.Println("value", value, " is string")
		return value.(string)
	case bool:
		Debug.Println("value", value, " is bool")
		if value == true {
			return "true"
		} else {
			return "false"
		}
	case map[string]interface{}:
		Debug.Println("value", value, " is map[string]interface{}")
		return GetDataOfSlice(value.(map[string]interface{}), resultpath, index+1)
	default:
		if b, err := json.Marshal(value); err == nil {
			Debug.Println("value", value, " is string")
			return string(b)
		} else {
			Error.Println("Error marshalling json", value, "Error:", err)
			panic("Error marshalling json")
		}
	}
}
