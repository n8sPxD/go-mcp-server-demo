package tools

var (
	supportTools = []ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Fetches the current weather for a given location.",
			InputSchema: ToolParameters{
				Type: "object",
				Properties: map[string]ToolParameterProperties{
					"city": {Type: "string", Description: "The city name to get weather for."},
				},
				Required: []string{"city"},
			},
		},
	}
)

// ToolParameterProperties 定义了工具参数的属性
type ToolParameterProperties struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ToolParameters 定义了工具的参数结构
type ToolParameters struct {
	Type       string                             `json:"type"` // 通常是 "object"
	Properties map[string]ToolParameterProperties `json:"properties"`
	Required   []string                           `json:"required,omitempty"`
}

// ToolDefinition 定义了一个工具
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema ToolParameters `json:"inputSchema"`
}

// ExecuteToolParams 是 tool/execute 请求的参数
type ExecuteToolParams struct {
	ToolName string         `json:"name"`
	Inputs   map[string]any `json:"arguments"`
}

// ExecuteToolResult 是 tool/execute 请求成功时的结果
type ExecuteToolResult struct {
	Content []map[string]any `json:"content"`
}

type ToolsMap map[string]ToolDefinition

func NewToolsMap() ToolsMap {
	toolsMap := make(ToolsMap)
	for _, tool := range supportTools {
		toolsMap.AddTool(tool)
	}
	return toolsMap
}

func (t ToolsMap) AddTool(tool ToolDefinition) {
	t[tool.Name] = tool
}

func (t ToolsMap) GetTool(name string) (ToolDefinition, bool) {
	tool, ok := t[name]
	return tool, ok
}
