package util

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	SymbolComma           = ","
	SymbolSingleQuotation = "'"
)

func StrAppend(strs ...string) string {
	var buffer bytes.Buffer
	for _, str := range strs {
		buffer.WriteString(str)
	}
	return buffer.String()
}

func JoinStrWithQuotation(separateSymbol string, quotation string, strs ...string) string {
	var buffer bytes.Buffer
	for _, code := range strs {
		buffer.WriteString(quotation)
		buffer.WriteString(code)
		buffer.WriteString(quotation)
		buffer.WriteString(separateSymbol)
	}
	buffer.Truncate(buffer.Len() - 1)
	return buffer.String()
}

func ToStrAry(input interface{}) []string {
	switch dtype := reflect.TypeOf(input).String(); dtype {
	case "string":
		str := input.(string)
		if str != "" {
			return []string{str}
		}
	case "[]string":
		return input.([]string)
	}
	return []string{}
}

func IntToFixStrLen(val int, length int) (string, error) {
	t := strconv.Itoa(val)
	valLen := len(t)
	if valLen > length {
		return "", errors.New(fmt.Sprintf("value %d is too long.", val))
	} else if valLen == length {
		return t, nil
	}

	returnStr := ""
	overLength := length - valLen
	for i := 0; i < overLength; i++ {
		returnStr = StrAppend(returnStr, "0")
	}
	return StrAppend(returnStr, t), nil
}
