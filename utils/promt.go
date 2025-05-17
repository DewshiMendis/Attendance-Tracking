package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Prompt(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
