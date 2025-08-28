package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

const (
	RAgentKey  = "ragent"
	ChatTplKey = "chat_tpl"
)

type state struct {
}

func buildChatGraph(ctx context.Context, cm model.ToolCallingChatModel) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	tpl := newChatTemplate()

	chatTools := GetTools()

	ragent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: chatTools,
		},
	})
	if err != nil {
		return nil, err
	}

	ragentGraph, ragentGraphOpts := ragent.ExportGraph()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{}
		}))

	_ = g.AddChatTemplateNode(ChatTplKey, tpl)
	_ = g.AddGraphNode(RAgentKey, ragentGraph, ragentGraphOpts...)

	_ = g.AddEdge(compose.START, ChatTplKey)
	_ = g.AddEdge(ChatTplKey, RAgentKey)
	_ = g.AddEdge(RAgentKey, compose.END)

	return g, nil
}
