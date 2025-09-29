package tools

import (
	"context"
	"fmt"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
)

func NewRAGTool(ctx context.Context) (tool.InvokableTool, *client.Client, error) {
	url := viper.GetString("mcp.rag.url")

	cli, err := client.NewStreamableHttpClient(url)
	if err != nil {
		return nil, nil, err
	}

	err = cli.Start(ctx)
	if err != nil {
		return nil, nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "rag-client",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, nil, err
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{
		Cli:          cli,
		ToolNameList: []string{"retrieve_documents_from_knowledge_base"},
	})
	if err != nil {
		return nil, nil, err
	}

	if len(tools) == 0 {
		return nil, nil, fmt.Errorf("no RAG tools found")
	}

	return tools[0].(tool.InvokableTool), cli, nil
}
