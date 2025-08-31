package agent

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

const (
	RAgentKey            = "ragent"
	RetrievalKey         = "retrieval"
	ChatTplKey           = "chat_tpl"
	RetrievalTplKey      = "retrieval_tpl"
	dataConvertLambdaKey = "data_convert_lambda"
)

type state struct {
	prompt  string
	history []*schema.Message
}

type RetrievalOutput struct {
	Retrieval bool   `json:"retrieval"`
	Docs      string `json:"docs"`
}

func buildChatGraph(ctx context.Context, cm model.ToolCallingChatModel, rcm model.ToolCallingChatModel) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	tpl := newChatTemplate()
	rtpl := newRetrievalTemplate()

	chatTools := GetTools()
	ragTools := GetRAGTools()

	retrievalRagent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: rcm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: ragTools,
		},
	})
	if err != nil {
		return nil, err
	}

	ragent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: chatTools,
		},
	})
	if err != nil {
		return nil, err
	}

	retrievalGraph, retrievalGraphOpts := retrievalRagent.ExportGraph()
	ragentGraph, ragentGraphOpts := ragent.ExportGraph()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{}
		}))

	_ = g.AddChatTemplateNode(RetrievalTplKey, rtpl, compose.WithStatePreHandler(retrievalPreHandler))
	_ = g.AddChatTemplateNode(ChatTplKey, tpl)
	_ = g.AddLambdaNode(dataConvertLambdaKey, compose.InvokableLambda(dataConvertLambda))
	_ = g.AddGraphNode(RetrievalKey, retrievalGraph, retrievalGraphOpts...)
	_ = g.AddGraphNode(RAgentKey, ragentGraph, ragentGraphOpts...)

	_ = g.AddEdge(compose.START, RetrievalTplKey)
	_ = g.AddEdge(RetrievalTplKey, RetrievalKey)
	_ = g.AddEdge(RetrievalKey, dataConvertLambdaKey)
	_ = g.AddEdge(dataConvertLambdaKey, ChatTplKey)
	_ = g.AddEdge(ChatTplKey, RAgentKey)
	_ = g.AddEdge(RAgentKey, compose.END)

	return g, nil
}

func retrievalPreHandler(ctx context.Context, input map[string]any, state *state) (map[string]any, error) {
	if historyValue, exists := input["history"]; exists {
		if history, ok := historyValue.([]*schema.Message); ok {
			state.history = history
		}
	}
	if promptValue, exists := input["prompt"]; exists {
		if prompt, ok := promptValue.(string); ok {
			state.prompt = prompt
		}
	}
	return input, nil
}

func dataConvertLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	retrievalOutput := RetrievalOutput{}

	if err := json.Unmarshal([]byte(input.Content), &retrievalOutput); err != nil {
		return nil, err
	}

	var result map[string]any

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		result = map[string]any{
			"history": state.history,
			"prompt":  state.prompt,
			"docs":    retrievalOutput.Docs,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}
