package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "DEBUG: MCP server started")

	file, err := os.OpenFile("/tmp/mcp_server_main_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file for main: %s", err)
	}
	logger := log.New(file, "[MCP Main] ", log.LstdFlags)    // 日志写入文件
	fmt.Fprintln(file, "DEBUG: MCP server started in main.") // 也写入文件

	// 使用标准输入输出进行通信
	// LSP/MCP 通常要求对输出进行缓冲，以便 Content-Length 等头部能正确写入
	// 这里直接使用 os.Stdout，但在 server.go 中的 sendResponse 确保了刷新
	// 对于读取，我们需要一个带缓冲的读取器
	stdinReader := bufio.NewReader(os.Stdin)
	stdoutWriter := bufio.NewWriter(os.Stdout) // 使用带缓冲的写入器

	server := NewMCPServer(stdinReader, stdoutWriter)
	logger.Println("MCP server instance created. Waiting for messages...")

	go func() {
		scanner := bufio.NewScanner(stdinReader)
		logger.Println("Scanner created. Using default line splitting. Entering scan loop...") // 更新日志

		for scanner.Scan() {
			messageBytes := scanner.Bytes()
			// messageBytes 现在是一行完整的JSON数据 (不包含换行符)
			if len(bytes.TrimSpace(messageBytes)) == 0 { // 可选：跳过空行
				logger.Println("Received an empty line, skipping.")
				continue
			}
			logger.Printf("DEBUG: Received message line: %s", string(messageBytes)) // 调试日志
			server.processMessage(messageBytes)
		}

		if err := scanner.Err(); err != nil {
			logger.Printf("Error reading from stdin: %v", err)
		}
		logger.Println("Stdin scanner finished.")
		// 如果输入结束，也应该关闭服务器
		close(server.shutdownSignal)
	}()

	// 等待服务器关闭信号
	<-server.shutdownSignal
	logger.Println("MCP server shut down gracefully.")
}
