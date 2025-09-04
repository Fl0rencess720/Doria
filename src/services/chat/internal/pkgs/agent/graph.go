package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/rag"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

const (
	RAgentKey            = "ragent"
	RetrievalAgentKey    = "retrieval"
	ChatTplKey           = "chat_tpl"
	RetrievalTplKey      = "retrieval_tpl"
	RetrievalLambdaKey   = "retrieval_lambda"
	dataConvertLambdaKey = "data_convert_lambda"
)

type state struct {
	history []*schema.Message
	prompt  string
	hr      *rag.HybridRetriever
}

type RetrievalAgentOutput struct {
	Retrieval bool   `json:"retrieval"`
	Query     string `json:"query"`
}

func buildChatGraph(ctx context.Context, cm model.ToolCallingChatModel, rcm model.ToolCallingChatModel, hr *rag.HybridRetriever) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	tpl := newChatTemplate()
	rtpl := newRetrievalTemplate()

	chatTools := GetChatTools()
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

	retrievalAgentGraph, retrievalAgentGraphOpts := retrievalRagent.ExportGraph()
	ragentGraph, ragentGraphOpts := ragent.ExportGraph()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{
				hr: hr,
			}
		}))

	_ = g.AddChatTemplateNode(RetrievalTplKey, rtpl, compose.WithStatePreHandler(retrievalPreHandler))
	_ = g.AddChatTemplateNode(ChatTplKey, tpl)
	_ = g.AddLambdaNode(dataConvertLambdaKey, compose.InvokableLambda(dataConvertLambda))
	_ = g.AddLambdaNode(RetrievalLambdaKey, compose.InvokableLambda(retrievalLambda))
	_ = g.AddGraphNode(RetrievalAgentKey, retrievalAgentGraph, retrievalAgentGraphOpts...)
	_ = g.AddGraphNode(RAgentKey, ragentGraph, ragentGraphOpts...)

	_ = g.AddEdge(compose.START, RetrievalTplKey)
	_ = g.AddEdge(RetrievalTplKey, RetrievalAgentKey)

	_ = g.AddBranch(RetrievalAgentKey, compose.NewGraphBranch(retrievalBranch, map[string]bool{
		RetrievalLambdaKey:   true,
		dataConvertLambdaKey: true,
	}))

	_ = g.AddEdge(RetrievalLambdaKey, ChatTplKey)
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

func retrievalBranch(ctx context.Context, input *schema.Message) (endNode string, err error) {
	retrievalAgebtOutput := &RetrievalAgentOutput{}

	if err := json.Unmarshal([]byte(input.Content), retrievalAgebtOutput); err != nil {
		return compose.END, err
	}

	if retrievalAgebtOutput.Retrieval {
		return RetrievalLambdaKey, nil
	}

	return dataConvertLambdaKey, nil
}

func dataConvertLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	var history []*schema.Message
	var prompt string

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		history = state.history
		prompt = state.prompt

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]any{
		"history": history,
		"prompt":  prompt,
		"docs":    "",
	}, nil
}

func retrievalLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	retrievalAgebtOutput := &RetrievalAgentOutput{}

	if err := json.Unmarshal([]byte(input.Content), retrievalAgebtOutput); err != nil {
		return nil, err
	}

	query := retrievalAgebtOutput.Query

	var hr *rag.HybridRetriever
	var history []*schema.Message
	var prompt string

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		hr = state.hr
		history = state.history
		prompt = state.prompt

		return nil
	}); err != nil {
		return nil, err
	}

	docs, err := hr.Retrieve(ctx, query)
	if err != nil {
		return nil, err
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

	return map[string]any{
		"history": history,
		"prompt":  prompt,
		"docs":    result,
	}, nil
}
