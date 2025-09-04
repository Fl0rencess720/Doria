package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/rag"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var chatTools []tool.BaseTool
var ragTools []tool.BaseTool

type RAGToolInput struct {
	Query string `json:"query" jsonschema:"description=查询文本,required"`
}

type RAGToolOutput struct {
	Result string `json:"result" jsonschema:"description=检索到的相关文档内容"`
}

func NewRAGTool(ctx context.Context) (tool.InvokableTool, error) {
	hr, err := rag.NewHybridRetriever(ctx)
	if err != nil {
		return nil, err
	}
	if hr == nil {
		return nil, fmt.Errorf("hybrid retriever is nil")
	}

	return utils.InferTool(
		"search_document",
		"根据查询文本检索相关文档内容",
		func(ctx context.Context, input *RAGToolInput) (*RAGToolOutput, error) {
			if input == nil {
				return &RAGToolOutput{
					Result: "输入参数为空",
				}, nil
			}

			docs, err := hr.Retrieve(ctx, input.Query)
			if err != nil {
				return &RAGToolOutput{
					Result: "检索失败: " + err.Error(),
				}, nil
			}

			var contents []string

			for _, doc := range docs {
				if doc != nil && doc.Content != "" {
					contents = append(contents, doc.Content)
				}
			}

			result := strings.Join(contents, "\n\n")
			if result == "" {
				result = "未找到相关文档"
			}

			return &RAGToolOutput{
				Result: result,
			}, nil
		},
	)
}

func tavilySearchTool(ctx context.Context) ([]tool.BaseTool, error) {
	cli, err := client.NewStreamableHttpClient(viper.GetString("tavily.URL") + viper.GetString("TAVILY_API_KEY"))
	if err != nil {
		return nil, err
	}
	if err := cli.Start(ctx); err != nil {
		return nil, err
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "bonfire-lit-client",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
	if err != nil {
		return nil, err
	}
	return tools, nil
}

func NewTools(ctx context.Context) {
	tavilyTools, err := tavilySearchTool(ctx)
	if err != nil {
		zap.L().Warn("tavily search tool init failed", zap.Error(err))
	}

	chatTools = append(chatTools, tavilyTools...)

	ragTool, err := NewRAGTool(ctx)
	if err != nil {
		zap.L().Warn("rag tool init failed", zap.Error(err))
		ragTool = nil
	}

	ragTools = append(ragTools, ragTool)
}

func GetChatTools() []tool.BaseTool {
	return chatTools
}

func GetRAGTools() []tool.BaseTool {
	return ragTools
}
