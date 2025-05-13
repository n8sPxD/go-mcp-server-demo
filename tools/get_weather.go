package tools

import (
	"fmt"

	"github.com/pkg/errors"
)

func GetWeather(city string) (*ExecuteToolResult, error) {
	if city == "" {
		return nil, errors.Wrap(
			errors.New("missing or invalid 'city' parameter for get_weather tool"),
			"missing or invalid 'city' parameter for get_weather tool",
		)
	}

	// 模拟天气获取
	weather := "Sunny"
	temperature := "25°C"
	if city == "London" {
		weather = "Cloudy"
		temperature = "15°C"
	} else if city == "Tokyo" {
		weather = "Rainy"
		temperature = "20°C"
	}

	// 将天气信息格式化为字符串
	weatherString := fmt.Sprintf("City: %s, Temperature: %s, Weather: %s", city, temperature, weather)

	// 构建符合客户端期望的 text content block
	textContentBlock := map[string]any{
		"type": "text",
		"text": weatherString,
	}

	// 将 text content block 包装在 Content 数组中
	result := ExecuteToolResult{
		Content: []map[string]any{textContentBlock},
	}

	return &result, nil
}
