package logger

import "fmt"

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
)

func LogWarn(str ...any) {
	fmt.Println(string(colorYellow), str, string(colorReset))
}

func LogError(str ...any) {
	fmt.Println(colorRed, str, colorReset)
}

func LogInfo(str ...any) {
	fmt.Println(string(colorGreen), str, string(colorReset))
}
