package tools

import (
	"fmt"

	"github.com/pkg/errors"
)

func Caculate(operation string, num1 float64, num2 float64) (float64, error) {
	switch operation {
	case "add":
		return num1 + num2, nil
	case "subtract":
		return num1 - num2, nil
	case "multiply":
		return num1 * num2, nil
	case "divide":
		return num1 / num2, nil
	}
	return 0, errors.New("invalid operation")
}

func ExecuteCaculate(operation string, num1 float64, num2 float64) (*ExecuteToolResult, error) {
	result, err := Caculate(operation, num1, num2)
	if err != nil {
		return nil, err
	}

	// 构建符合客户端期望的 text content block
	textContentBlock := map[string]any{
		"type": "text",
		"text": fmt.Sprintf("The result of %s %f and %f is %f", operation, num1, num2, result),
	}

	return &ExecuteToolResult{
		Content: []map[string]any{textContentBlock},
	}, nil
}
