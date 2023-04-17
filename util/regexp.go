package util

import (
	"regexp"
)

func MatchString(b []byte, expr string) bool {
	reg, err := regexp.Compile(expr)
	if err != nil {
		panic(err)
	}

	if reg.Match(b) {
		return true
	}
	//for {
	//	line, _, err := b.ReadLine()
	//	if err != nil {
	//		break
	//	}
	//	if reg.Match(line) {
	//		return true
	//	}
	//}
	return false
}
