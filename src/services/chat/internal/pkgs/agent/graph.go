package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Fl0rencess720/Bonfire-Lit/src/common/rag"
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

func buildChatGraph(ctx context.Context, mainCM model.ToolCallingChatModel, intentCM model.ToolCallingChatModel, hr *rag.HybridRetriever) (*compose.Graph[map[string]any, *schema.Message], error) {
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
				hr: hr, // 将HybridRetriever作为state的一部分传递
			}
		}))
	// 3. IntentPromptTplKey: 意图识别的提示模板节点。
	// 原名：RetrievalTplKey (更名为 IntentPromptTplKey，明确其用于意图识别)
	_ = g.AddChatTemplateNode(IntentPromptTplKey, intentTpl, compose.WithStatePreHandler(inputContextLoader)) // preHandler也需要更名
	// 4. ResponsePromptTplKey: 最终回复的提示模板节点。
	// 原名：ChatTplKey (更名为 ResponsePromptTplKey，明确其用于生成最终回复)
	_ = g.AddChatTemplateNode(ResponsePromptTplKey, responseTpl)
	// 5. NoRAGDataPreparerLambdaKey: 当意图识别判断不需要RAG时，准备数据跳过检索。
	// 原名：dataConvertLambdaKey (不够具体，应该明确它什么情况下转换数据以及转换的目的)
	_ = g.AddLambdaNode(NoRAGDataPreparerLambdaKey, compose.InvokableLambda(noRAGDataPreparerLambda))
	// 6. DocumentRetrieverLambdaKey: 执行文档检索并格式化结果的Lambda节点。
	// 原名：RetrievalLambdaKey (更名为 DocumentRetrieverLambdaKey，明确其执行RAG Retrieval)
	_ = g.AddLambdaNode(RetrieverLambdaKey, compose.InvokableLambda(retrievalLambda))
	// 7. IntentDetectorAgentKey: 代理节点，执行意图检测。
	// 原名：RetrievalAgentKey (更名为 IntentDetectorAgentKey，明确其是负责意图判断的Agent)
	_ = g.AddGraphNode(IntentDetectorAgentKey, intentDetectorAgentGraph, intentDetectorAgentGraphOpts...)
	// 8. MainChatAgentKey: 代理节点，执行主对话和最终回复生成。
	// 原名：RAgentKey (更名为 MainChatAgentKey，明确其是核心的对话Agent)
	_ = g.AddGraphNode(MainChatAgentKey, mainChatAgentGraph, mainChatAgentGraphOpts...)
	// =================================== 图的边和分支定义 ===================================
	// 流程开始，首先进入意图识别的提示模板。
	_ = g.AddEdge(compose.START, IntentPromptTplKey)
	// 意图识别提示模板的输出作为输入传递给意图识别代理。
	_ = g.AddEdge(IntentPromptTplKey, IntentDetectorAgentKey)
	// 意图识别代理根据判断结果进行分支。
	// 命名：RAGDecisionBranch，明确它是RAG决策的分支。
	_ = g.AddBranch(IntentDetectorAgentKey, compose.NewGraphBranch(ragDecisionBranch, map[string]bool{
		RetrieverLambdaKey:         true,
		NoRAGDataPreparerLambdaKey: true,
	}))
	// 文档检索Lambda的输出（包含文档）作为输入给最终回复提示模板。
	_ = g.AddEdge(RetrieverLambdaKey, ResponsePromptTplKey)
	// 跳过检索的Lambda的输出（不含文档）作为输入给最终回复提示模板。
	_ = g.AddEdge(NoRAGDataPreparerLambdaKey, ResponsePromptTplKey)
	// 最终回复提示模板的输出作为输入给主对话代理。
	_ = g.AddEdge(ResponsePromptTplKey, MainChatAgentKey)
	// 主对话代理的输出即为整个图的最终输出。
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
