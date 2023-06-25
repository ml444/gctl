package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var regStatusCode = regexp.MustCompile(`^.*?@status_code:\s*(\d+)\s*$`)

func Concat(args ...string) string {
	var strBuilder = &strings.Builder{}
	for _, s := range args {
		strBuilder.WriteString(s)
	}
	return strBuilder.String()
}

func GetStatusCodeFromComment(comment string) (int32, error) {
	if !strings.Contains(comment, "@status_code:") {
		return 0, nil
	}
	matchResult := regStatusCode.FindStringSubmatch(comment)
	if len(matchResult) != 2 {
		return 0, nil
	}
	v, err := strconv.ParseInt(matchResult[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func ToUpperFirst(s string) string {
	if s == "" {
		return s
	}
	if d := s[0]; d >= 'a' && d <= 'z' {
		return string(d-32) + s[1:]
	}
	return s
}

func ToLowerFirst(s string) string {
	if s == "" {
		return s
	}
	if d := s[0]; d >= 'A' && d <= 'Z' {
		return string(d+32) + s[1:]
	}
	return s
}

func CamelToSnake(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func SnakeToCamel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func ToInt(v interface{}) int {
	switch vv := v.(type) {
	case int:
		return vv
	case float64:
		return int(vv)
	case float32:
		return int(vv)
	default:
		panic(fmt.Sprintf("type %v not expected", vv))
	}
}
