package service

import (
	"context"
	"encoding/json"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type RAGMCPService struct {
	httpServer *server.StreamableHTTPServer
	uc         *biz.RAGUseCase
}

func NewRAGMCPServer(ragUseCase *biz.RAGUseCase) *RAGMCPService {
	s := server.NewMCPServer(
		"RAG Service",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	retrieveTool := mcp.NewTool("retrieve_documents_from_knowledge_base",
		mcp.WithDescription("从知识库中搜索并检索相关文档来回答用户问题。当您需要对话历史中不可用的特定信息、事实或上下文时，请使用此工具。"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("从用户问题和对话上下文中提取的独立、具体的搜索查询。查询应该包含丰富的关键词，准确表示用户寻求的核心信息。例如，如果用户问'我们对新登录页面做了什么决定？'，一个好的查询是'新登录页面设计决策'，而不是'我们对新登录页面做了什么决定'。"),
		),
	)

	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp/rag"),
	)

	service := &RAGMCPService{
		httpServer: httpServer,
		uc:         ragUseCase,
	}

	s.AddTool(retrieveTool, service.Retrieve)

	return service
}

func (s *RAGMCPService) Start() {
	go func() {
		if err := s.httpServer.Start(":8082"); err != nil {
			zap.L().Error("Failed to start HTTP server", zap.Error(err))
		}
	}()
}

func (s *RAGMCPService) Retrieve(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		zap.L().Error("Failed to get query from request", zap.Error(err))
		return nil, err
	}

	documents, err := s.uc.Retrieve(ctx, query)
	if err != nil {
		zap.L().Error("Failed to retrieve documents", zap.Error(err))
		return nil, err
	}

	jsonDocuments, err := json.Marshal(documents)
	if err != nil {
		zap.L().Error("Failed to marshal documents", zap.Error(err))
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonDocuments)), nil
}
