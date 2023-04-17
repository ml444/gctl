package util

func Add(args ...int) int {
	var sum int
	if l := len(args); l == 0 {
		return 0
	} else if l == 1 {
		return ToInt(args[0])
	}
	sum, args = args[0], args[1:]
	for _, v := range args {
		sum += v
	}
	return sum
}
