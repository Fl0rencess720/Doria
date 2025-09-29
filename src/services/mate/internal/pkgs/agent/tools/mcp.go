package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"go.uber.org/zap"
)

type MCPManager struct {
	ragTool tool.InvokableTool
	clients []*client.Client
}

func NewMCPManager(ctx context.Context) (*MCPManager, error) {
	ragTool, cli, err := NewRAGTool(ctx)
	if err != nil {
		return nil, err
	}

	return &MCPManager{
		ragTool: ragTool,
		clients: []*client.Client{cli},
	}, nil
}

func (m *MCPManager) GetRAGTool() tool.InvokableTool {
	return m.ragTool
}

func (m *MCPManager) Close() {
	for _, cli := range m.clients {
		if err := cli.Close(); err != nil {
			zap.L().Error("Error closing MCP client", zap.Error(err))
		}
	}
}
