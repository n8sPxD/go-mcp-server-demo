package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

func GetWeather(location string) (*ExecuteToolResult, error) {
	if location == "" {
		return nil, errors.Wrap(
			errors.New("missing or invalid 'location' parameter for get_weather tool"),
			"missing or invalid 'location' parameter for get_weather tool",
		)
	}

	// 获取天气信息
	weather, err := getWeather(location)
	if err != nil {
		return nil, err
	}

	// 将天气信息格式化为字符串
	weatherString := fmt.Sprintf("Location: %s, Weather: %s, Temperature: %.1f°C", location, weather.Weather, weather.TempC)

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

func getWeather(location string) (*CommonWeatherResponse, error) {
	weatherGetter, err := NewWeatherAPIWeatherGetter()
	if err != nil {
		return nil, err
	}
	return weatherGetter.GetWeather(location)
}

type WeatherGetter interface {
	GetWeather(location string) (*CommonWeatherResponse, error)
}

type CommonWeatherResponse struct {
	Location string  `json:"location"`
	TempC    float64 `json:"temp_c"`
	Weather  string  `json:"weather"`
}

type WeatherAPIWeatherGetter struct {
	apiKey string
}

func NewWeatherAPIWeatherGetter() (WeatherGetter, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("WEATHER_API_KEY is not set, please set it in the environment variables")
	}
	return &WeatherAPIWeatherGetter{apiKey: apiKey}, nil
}

type WeatherAPIParams struct {
	APIKey string `url:"key"`
	Q      string `url:"q"`   // location
	AQI    string `url:"aqi"` // "yes" or "no"
}

type WeatherAPIResponse struct {
	Location struct {
		Name    string  `json:"name"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		TempF     float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func (w *WeatherAPIWeatherGetter) GetWeather(location string) (*CommonWeatherResponse, error) {
	// 获取天气信息
	params := &WeatherAPIParams{
		APIKey: w.apiKey,
		Q:      location,
		AQI:    "no",
	}

	url, err := url.Parse("https://api.weatherapi.com/v1/current.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse weather API URL")
	}

	q := url.Query()
	q.Add("key", params.APIKey)
	q.Add("q", params.Q)
	q.Add("aqi", params.AQI)
	url.RawQuery = q.Encode()

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get weather API response")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read weather API response body")
	}

	var weatherAPIResponse WeatherAPIResponse
	if err := json.Unmarshal(body, &weatherAPIResponse); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal weather API response")
	}

	commonWeatherResponse := CommonWeatherResponse{
		Location: weatherAPIResponse.Location.Name,
		TempC:    weatherAPIResponse.Current.TempC,
		Weather:  weatherAPIResponse.Current.Condition.Text,
	}

	return &commonWeatherResponse, nil
}
