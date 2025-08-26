package agent

import (
	"context"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
)

func tavilySearchTool(ctx context.Context) ([]tool.BaseTool, error) {
	cli, err := client.NewSSEMCPClient(viper.GetString("tavily.URL") + viper.GetString("TAVILY_API_KEY"))
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

func getTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{}
	tavilyTools, err := tavilySearchTool(ctx)
	if err != nil {
		return nil, err
	}
	return append(tools, tavilyTools...), nil
	return tools, nil
}
