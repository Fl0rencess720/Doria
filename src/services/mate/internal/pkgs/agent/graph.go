package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

const (
	MainChatAgentKey           = "main_chat_agent"
	IntentDetectorAgentKey     = "intent_detector_agent"
	IntentPromptTplKey         = "intent_prompt_tpl"
	ResponsePromptTplKey       = "response_prompt_tpl"
	RetrieverLambdaKey         = "retriever_lambda"
	NoRAGDataPreparerLambdaKey = "no_rag_data_preparer_lambda"
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

func buildChatGraph(ctx context.Context, mainCM model.ToolCallingChatModel,
	intentCM model.ToolCallingChatModel, hr *rag.HybridRetriever) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	responseTpl := newResponseTemplate()
	intentTpl := newIntentTemplate()

	chatTools := GetChatTools()

	intentDetectorAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: intentCM,
	})
	if err != nil {
		return nil, err
	}

	mainChatAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: mainCM,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: chatTools,
		},
	})
	if err != nil {
		return nil, err
	}

	intentDetectorAgentGraph, intentDetectorAgentGraphOpts := intentDetectorAgent.ExportGraph()
	mainChatAgentGraph, mainChatAgentGraphOpts := mainChatAgent.ExportGraph()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{
				hr: hr,
			}
		}))

	_ = g.AddChatTemplateNode(IntentPromptTplKey, intentTpl, compose.WithStatePreHandler(inputContextLoader))
	_ = g.AddChatTemplateNode(ResponsePromptTplKey, responseTpl)

	_ = g.AddLambdaNode(NoRAGDataPreparerLambdaKey, compose.InvokableLambda(noRAGDataPreparerLambda))
	_ = g.AddLambdaNode(RetrieverLambdaKey, compose.InvokableLambda(retrievalLambda))

	_ = g.AddGraphNode(IntentDetectorAgentKey, intentDetectorAgentGraph, intentDetectorAgentGraphOpts...)
	_ = g.AddGraphNode(MainChatAgentKey, mainChatAgentGraph, mainChatAgentGraphOpts...)

	_ = g.AddEdge(compose.START, IntentPromptTplKey)
	_ = g.AddEdge(IntentPromptTplKey, IntentDetectorAgentKey)
	_ = g.AddBranch(IntentDetectorAgentKey, compose.NewGraphBranch(ragDecisionBranch, map[string]bool{
		RetrieverLambdaKey:         true,
		NoRAGDataPreparerLambdaKey: true,
	}))
	_ = g.AddEdge(RetrieverLambdaKey, ResponsePromptTplKey)
	_ = g.AddEdge(NoRAGDataPreparerLambdaKey, ResponsePromptTplKey)
	_ = g.AddEdge(ResponsePromptTplKey, MainChatAgentKey)
	_ = g.AddEdge(MainChatAgentKey, compose.END)

	return g, nil
}

func inputContextLoader(ctx context.Context, input map[string]any, state *state) (map[string]any, error) {
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

func ragDecisionBranch(ctx context.Context, input *schema.Message) (endNode string, err error) {
	retrievalAgebtOutput := &RetrievalAgentOutput{}

	if err := json.Unmarshal([]byte(input.Content), retrievalAgebtOutput); err != nil {
		zap.L().Error("failed to unmarshal retrieval agent output", zap.Error(err))
		return NoRAGDataPreparerLambdaKey, nil
	}

	if retrievalAgebtOutput.Retrieval {
		return RetrieverLambdaKey, nil
	}

	return NoRAGDataPreparerLambdaKey, nil
}

func noRAGDataPreparerLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
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

	for i, doc := range docs {
		if doc != nil && doc.Content != "" {
			contents = append(contents, fmt.Sprintf("文档片段 %d:\n%s", i+1, doc.Content))
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
