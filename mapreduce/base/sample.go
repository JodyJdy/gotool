package base

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// MapFunc 词频统计 map reduce 代码
func MapFunc(pair Pair) (res []Pair) {
	value := pair.Val2Str()
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	words := strings.FieldsFunc(value, f)
	for _, key := range words {
		res = append(res, Pair{key, "1"})
	}
	return res
}

func ReduceFunc(key string, vals []Value) string {
	count := 0
	for _, value := range vals {
		str := reflect.ValueOf(value).Interface().(string)
		num, _ := strconv.ParseInt(str, 10, 64)
		count = count + int(num)
	}
	return strconv.Itoa(count)
}
