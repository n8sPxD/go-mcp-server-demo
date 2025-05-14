package server

import (
	"encoding/json"

	"github.com/n8sPxD/mcp-server-demo/tools"
)

// JSONRPCVersion 是固定的 JSON-RPC 版本号
const JSONRPCVersion = "2.0"

// BaseMessage 是所有 JSON-RPC 消息的基础结构
type BaseMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"` // 指针类型，因为请求有ID，通知没有ID，响应ID匹配请求ID
}

// RequestMessage 代表一个 JSON-RPC 请求
type RequestMessage struct {
	BaseMessage
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ResponseMessage 代表一个 JSON-RPC 响应
type ResponseMessage struct {
	BaseMessage
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorObject    `json:"error,omitempty"`
}

// NotificationMessage 代表一个 JSON-RPC 通知
type NotificationMessage struct {
	JSONRPC string          `json:"jsonrpc"` // 通知没有ID，所以不嵌入BaseMessage
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// ErrorObject 代表 JSON-RPC 错误对象
type ErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC Error Codes
const (
	ParseErrorCode     = -32700
	InvalidRequestCode = -32600
	MethodNotFoundCode = -32601
	InvalidParamsCode  = -32602
	InternalErrorCode  = -32603
)

// ClientInfo 包含客户端的信息
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// InitializeParams 是 initialize 请求的参数
type InitializeParams struct {
	ProtocolVersion  *string           `json:"protocolVersion,omitempty"`
	ProcessID        *int              `json:"processId"` // 可以为 null
	ClientInfo       *ClientInfo       `json:"clientInfo,omitempty"`
	Capabilities     json.RawMessage   `json:"capabilities,omitempty"`     // 客户端能力，暂时不详细解析
	Trace            string            `json:"trace,omitempty"`            // "off", "messages", "verbose"
	RootURI          *string           `json:"rootUri"`                    // 可以为 null
	WorkspaceFolders []json.RawMessage `json:"workspaceFolders,omitempty"` // 暂时不详细解析
}

// ServerInfo 包含服务器的信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ServerCapabilities 定义了服务器的能力
type ServerCapabilities struct {
	Tools map[string]tools.ToolDefinition `json:"tools,omitempty"`
	// 可以添加其他能力，例如 textDocumentSync, completionProvider 等
}

// InitializeResult 是 initialize 请求成功时的结果
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

// InitializedParams 是 initialized 通知的参数 (通常为空)
type InitializedParams struct{}

// ShutdownParams 是 shutdown 请求的参数 (通常为空或null)
type ShutdownParams struct{}

// ExitParams 是 exit 通知的参数 (通常为空或null)
type ExitParams struct{}

// ListToolsParams 是 tools/list 请求的参数 (通常为空，但为完整性添加)
type ListToolsParams struct{}

// ListToolsResult 是 tools/list 请求成功时的结果
type ListToolsResult struct {
	Tools []tools.ToolDefinition `json:"tools"`
}
