package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/n8sPxD/mcp-server-demo/tools"
)

// MCPServer 定义了 MCP 服务器的状态和能力
type MCPServer struct {
	reader      io.Reader
	writer      io.Writer
	logger      *log.Logger
	file        *os.File
	tools       tools.ToolsMap
	initialized bool

	ShutdownSignal chan struct{} // 用于通知主循环服务器已关闭
}

func NewMCPServer(reader io.Reader, writer io.Writer, file *os.File) *MCPServer {
	return &MCPServer{
		reader:         reader,
		writer:         writer,
		file:           file,
		logger:         log.New(file, "[MCP Server] ", log.LstdFlags),
		tools:          make(map[string]tools.ToolDefinition),
		initialized:    false,
		ShutdownSignal: make(chan struct{}),
	}
}

// sendResponse 发送 JSON-RPC 响应
func (s *MCPServer) sendResponse(id *json.RawMessage, result any, err *ErrorObject) {
	response := ResponseMessage{
		BaseMessage: BaseMessage{
			JSONRPC: JSONRPCVersion,
			ID:      id,
		},
	}
	if err != nil {
		response.Error = err
	} else {
		resultBytes, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			s.logger.Printf("Error marshalling result: %v", marshalErr)
			// Fallback to sending an internal error if marshalling the actual result fails
			response.Error = &ErrorObject{Code: InternalErrorCode, Message: "Error marshalling result"}
			response.Result = nil // Clear any potentially partially set result
		} else {
			response.Result = resultBytes
		}
	}

	responseBytes, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		s.logger.Printf("Error marshalling response: %v", marshalErr)
		// Cannot send a response if we can't marshal the response itself.
		// Log and potentially panic or exit, depending on desired robustness.
		return
	}

	// 打印格式化后响应
	prettyResponse, _ := json.MarshalIndent(response, "", "  ")
	s.logger.Printf("Sending formatted response: %s\n", string(prettyResponse))

	// 直接写入响应体
	if _, writeErr := s.writer.Write(responseBytes); writeErr != nil {
		s.logger.Printf("Error writing body: %v", writeErr)
		return // 如果写入响应体失败，也应该返回
	}
	// 在响应体后写入换行符
	if _, writeErr := s.writer.Write([]byte("\n")); writeErr != nil {
		s.logger.Printf("Error writing newline: %v", writeErr)
		return
	}

	if flusher, ok := s.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			s.logger.Printf("Error flushing writer: %v", err)
		}
	}
}

// sendNotification 发送 JSON-RPC 通知 (这里未使用，但可以按需添加)
/*
func (s *MCPServer) sendNotification(method string, params any) {
	notification := NotificationMessage{
		JSONRPC: JSONRPCVersion,
		Method:  method,
	}
	if params != nil {
		paramsBytes, _ := json.Marshal(params)
		notification.Params = paramsBytes
	}
	notificationBytes, _ := json.Marshal(notification)
	s.logger.Printf("Sending notification: %s\n", string(notificationBytes))

	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", len(notificationBytes))
	s.writer.Write([]byte(header))
	s.writer.Write(notificationBytes)
	if flusher, ok := s.writer.(interface{ Flush() error }); ok {
		flusher.Flush()
	}
}
*/

// handleInitialize 处理 initialize 请求
func (s *MCPServer) handleInitialize(req RequestMessage) {
	var params InitializeParams // <--- 用于解析请求参数
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.logger.Printf("Error unmarshalling initialize params: %v", err)
		s.sendResponse(req.ID, nil, &ErrorObject{Code: InvalidParamsCode, Message: "Invalid params for initialize"})
		return
	}
	s.logger.Printf("Initialize params: %+v\n", params) // 打印解析后的参数

	serverInfo := ServerInfo{
		Name:    "mcp-go-weather-server",
		Version: "0.0.1",
	}

	// 将 Tools 从数组改为 map
	toolsMap := tools.NewToolsMap()
	s.tools = toolsMap

	capabilities := ServerCapabilities{
		Tools: s.tools, // <--- 使用 s.tools
	}

	// 从客户端参数中获取 protocolVersion，如果不存在则使用默认值
	clientProtocolVersion := ""
	if params.ProtocolVersion != nil {
		clientProtocolVersion = *params.ProtocolVersion
	}
	s.logger.Printf("Client requested protocol version: %s", clientProtocolVersion)

	result := InitializeResult{
		ProtocolVersion: clientProtocolVersion, // <--- 设置 ProtocolVersion
		ServerInfo:      serverInfo,
		Capabilities:    capabilities,
	}
	s.sendResponse(req.ID, result, nil)
}

// handleInitialized 处理 initialized 通知
func (s *MCPServer) handleInitialized(notif NotificationMessage) {
	s.logger.Println("Server initialized by client.")
	s.initialized = true
	// 可以在这里执行初始化后的操作
}

// handleShutdown 处理 shutdown 请求
func (s *MCPServer) handleShutdown(req RequestMessage) {
	s.logger.Println("Shutdown request received.")
	s.sendResponse(req.ID, nil, nil) // 回复空结果
	// 准备关闭，但不立即退出，等待 exit 通知
}

// handleExit 处理 exit 通知
func (s *MCPServer) handleExit(notif NotificationMessage) {
	s.logger.Println("Exit notification received. Server shutting down.")
	close(s.ShutdownSignal) // 发送关闭信号
}

// handleExecuteTool 处理 tools/call 请求
func (s *MCPServer) handleExecuteTool(req RequestMessage) {
	if !s.initialized {
		s.sendResponse(req.ID, nil, &ErrorObject{Code: InternalErrorCode, Message: "Server not initialized"})
		return
	}

	var params tools.ExecuteToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendResponse(req.ID, nil, &ErrorObject{Code: InvalidParamsCode, Message: "Invalid params for tools/call"})
		return
	}

	s.logger.Printf("Executing tool: %s with inputs: %+v\n", params.ToolName, params.Inputs)

	if toolFunc, ok := tools.ToolFuncMap[params.ToolName]; ok {
		content, err := toolFunc(params.Inputs)
		if err != nil {
			s.sendResponse(req.ID, nil, &ErrorObject{Code: InternalErrorCode, Message: err.Error()})
			return
		}
		s.sendResponse(req.ID, content, nil)
	} else {
		s.sendResponse(req.ID, nil, &ErrorObject{Code: MethodNotFoundCode, Message: fmt.Sprintf("Tool '%s' not found", params.ToolName)})
	}
}

// handleListTools 处理 tools/list 请求
func (s *MCPServer) handleListTools(req RequestMessage) {
	if !s.initialized {
		s.sendResponse(req.ID, nil, &ErrorObject{Code: InternalErrorCode, Message: "Server not initialized"})
		return
	}

	s.logger.Println("ListTools request received.")

	// 将 s.tools (map) 转换为 []ToolDefinition
	var toolsArray []tools.ToolDefinition
	for _, toolDef := range s.tools {
		toolsArray = append(toolsArray, toolDef)
	}

	result := ListToolsResult{
		Tools: toolsArray, // <--- 使用转换后的数组
	}
	s.sendResponse(req.ID, result, nil)
}

// processMessage 解析并处理单个消息
func (s *MCPServer) ProcessMessage(rawMessage []byte) {
	// 打印格式化后的消息
	var tempMarshalMap map[string]interface{}
	if err := json.Unmarshal(rawMessage, &tempMarshalMap); err == nil {
		prettyMessage, _ := json.MarshalIndent(tempMarshalMap, "", "  ")
		s.logger.Printf("DEBUG: Received message line: %s", string(prettyMessage)) // 调试日志
	} else {
		s.logger.Printf("DEBUG: Received message line: %s", string(rawMessage)) // 调试日志
	}

	// 首先尝试解析基本结构，以判断是请求还是通知 (通过有无ID)
	var base BaseMessage
	if err := json.Unmarshal(rawMessage, &base); err != nil {
		// 如果连基本结构都无法解析，记录错误。无法确定ID，无法响应。
		s.logger.Printf("Failed to parse base JSON message: %v. Raw: %s\n", err, string(rawMessage))
		return
	}

	if base.ID != nil { // 有 ID，说明是请求 (或者是我们不期望从客户端收到的响应)
		var req RequestMessage
		// 再次解析为完整的 RequestMessage 结构
		if err := json.Unmarshal(rawMessage, &req); err == nil && req.Method != "" {
			s.logger.Printf("Parsed as Request: ID=%s, Method=%s\n", string(*req.ID), req.Method)
			// 打印格式化的请求内容
			if s.logger != nil {
				prettyReq, _ := json.MarshalIndent(req, "", "  ")
				s.logger.Printf("Received formatted request: %s\n", string(prettyReq))
			}

			switch req.Method {
			case "initialize":
				s.handleInitialize(req)
			case "shutdown":
				s.handleShutdown(req)
			case "tools/call":
				s.handleExecuteTool(req)
			case "tools/list":
				s.handleListTools(req)
			default:
				s.logger.Printf("Unknown request method: %s\n", req.Method)
				s.sendResponse(req.ID, nil, &ErrorObject{Code: MethodNotFoundCode, Message: "Method not found: " + req.Method})
			}
		} else {
			// 有ID但无法解析为有效请求 (例如，缺少method字段)
			s.logger.Printf("Received message with ID that is not a valid request structure. Raw: %s, Parse Err: %v\n", string(rawMessage), err)
			s.sendResponse(base.ID, nil, &ErrorObject{Code: InvalidRequestCode, Message: "Invalid Request"})
		}
	} else { // 没有 ID，说明是通知
		var notif NotificationMessage
		if err := json.Unmarshal(rawMessage, &notif); err == nil && notif.Method != "" {
			s.logger.Printf("Parsed as Notification: Method=%s\n", notif.Method)
			// 打印格式化的通知内容
			if s.logger != nil {
				prettyNotif, _ := json.MarshalIndent(notif, "", "  ")
				s.logger.Printf("Received formatted notification: %s\n", string(prettyNotif))
			}

			switch notif.Method {
			case "initialized", "notifications/initialized":
				s.handleInitialized(notif)
			case "exit":
				s.handleExit(notif)
			// 可以添加其他通知处理，例如 $/cancelRequest
			default:
				s.logger.Printf("Unknown notification method: %s\n", notif.Method)
			}
		} else {
			// 没有ID，并且无法解析为有效的通知结构
			s.logger.Printf("Failed to parse message as Notification (and it had no ID). Raw: %s, Parse Err: %v\n", string(rawMessage), err)
			// 对于无法解析的通知，通常不发送响应
		}
	}
}
