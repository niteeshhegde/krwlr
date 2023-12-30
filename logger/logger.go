package logger

import "fmt"

var colorReset string = "\033[0m"
var colorRed string = "\033[31m"
var colorGreen string = "\033[32m"
var colorYellow string = "\033[33m"

func LogWarn(str ...any) {
	fmt.Println(string(colorYellow), str, string(colorReset))
}

func LogError(str ...any) {
	fmt.Println(colorRed, str, colorReset)
}

func LogInfo(str ...any) {
	fmt.Println(string(colorGreen), str, string(colorReset))
}
