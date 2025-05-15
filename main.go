package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/n8sPxD/mcp-server-demo/server"
)

func main() {
	fmt.Fprintln(os.Stderr, "DEBUG: MCP server started")

	// 日志写入文件
	file, err := os.OpenFile("/tmp/mcp_server_main_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file for main: %s", err)
	}

	logger := log.New(file, "[MCP Main] ", log.LstdFlags)    // 日志写入文件
	fmt.Fprintln(file, "DEBUG: MCP server started in main.") // 也写入文件

	stdinReader := bufio.NewReader(os.Stdin)
	stdoutWriter := bufio.NewWriter(os.Stdout)

	server := server.NewMCPServer(stdinReader, stdoutWriter, file)
	logger.Println("MCP server instance created. Waiting for messages...")

	go func() {
		scanner := bufio.NewScanner(stdinReader)
		logger.Println("Scanner created. Using default line splitting. Entering scan loop...") // 更新日志

		for scanner.Scan() {
			messageBytes := scanner.Bytes()
			if len(bytes.TrimSpace(messageBytes)) == 0 {
				logger.Println("Received an empty line, skipping.")
				continue
			}

			server.ProcessMessage(messageBytes)
		}

		if err := scanner.Err(); err != nil {
			logger.Printf("Error reading from stdin: %v", err)
		}
		logger.Println("Stdin scanner finished.")
		// 如果输入结束，也应该关闭服务器
		close(server.ShutdownSignal)
	}()

	// 等待服务器关闭信号
	<-server.ShutdownSignal
	logger.Println("MCP server shut down gracefully.")
}
