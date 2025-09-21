package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	GuidelineProposerPromptTplKey = "guideline_proposer_prompt"
	ToolCallerPromptTplKey        = "tool_caller_prompt"
	ObserverPomptTplKey           = "observer_prompt"
	DoriaPromptTplKey             = "doria_prompt"

	GuidelineProposerChatModelKey = "guideline_proposer_chat_model"
	ToolCallerChatModelKey        = "tool_caller_chat_model"
	ObserverChatModelKey          = "observer_chat_model"
	DoriaChatModelKey             = "doria_chat_model"

	ActiveGuidelinesLambdaKey     = "active_guidelines_lambda"
	ToolCallingLambdaKey          = "tool_calling_lambda"
	ConvertObserverOuputLambdaKey = "convert_observer_output_lambda"
)

type state struct {
	history []*schema.Message
	prompt  string

	hr *rag.HybridRetriever

	guidelines       []*Guideline
	guidelinesString string

	activeGuidelines       []*Guideline
	activeGuidelinesString string
	toolOutput             string

	epoch int
}

type InputPayload struct {
	GuidelineEvaluations []*GuidelineEvaluation `json:"guideline_evaluations"`
}

type ObserverOutput struct {
	Toward bool   `json:"toward"`
	Reason string `json:"reason"`
}

func buildChatGraph(_ context.Context, mainCM model.ToolCallingChatModel,
	jsonCM model.ToolCallingChatModel, hr *rag.HybridRetriever) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	guidelineProposerTpl := newGuidelineProposerResponseTemplate()
	toolCallerTpl := newToolCallerResponseTemplate()
	observerTpl := newObserverResponseTemplate()
	doriaTpl := newDoriaResponseTemplate()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{
				hr: hr,
			}
		}))

	_ = g.AddChatTemplateNode(GuidelineProposerPromptTplKey, guidelineProposerTpl, compose.WithStatePreHandler(saveInputToState))
	_ = g.AddChatTemplateNode(ToolCallerPromptTplKey, toolCallerTpl)
	_ = g.AddChatTemplateNode(ObserverPomptTplKey, observerTpl)
	_ = g.AddChatTemplateNode(DoriaPromptTplKey, doriaTpl)

	_ = g.AddChatModelNode(GuidelineProposerChatModelKey, jsonCM)
	_ = g.AddChatModelNode(ToolCallerChatModelKey, jsonCM)
	_ = g.AddChatModelNode(ObserverChatModelKey, jsonCM)
	_ = g.AddChatModelNode(DoriaChatModelKey, mainCM)

	_ = g.AddLambdaNode(ActiveGuidelinesLambdaKey, compose.InvokableLambda(activeGuidelinesLambda))
	_ = g.AddLambdaNode(ToolCallingLambdaKey, compose.InvokableLambda(toolCallingLambda))
	_ = g.AddLambdaNode(ConvertObserverOuputLambdaKey, compose.InvokableLambda(convertObserverOutputLambda))

	_ = g.AddEdge(compose.START, GuidelineProposerPromptTplKey)
	_ = g.AddEdge(GuidelineProposerPromptTplKey, GuidelineProposerChatModelKey)
	_ = g.AddEdge(GuidelineProposerChatModelKey, ActiveGuidelinesLambdaKey)

	_ = g.AddBranch(ActiveGuidelinesLambdaKey, compose.NewGraphBranch(toolCallingDecisionBranch, map[string]bool{
		ToolCallerPromptTplKey: true,
		DoriaPromptTplKey:      true,
	}))

	_ = g.AddEdge(ToolCallerPromptTplKey, ToolCallerChatModelKey)
	_ = g.AddEdge(ToolCallerChatModelKey, ToolCallingLambdaKey)
	_ = g.AddEdge(ToolCallingLambdaKey, ObserverPomptTplKey)
	_ = g.AddEdge(ObserverPomptTplKey, ObserverChatModelKey)
	_ = g.AddEdge(ObserverChatModelKey, ConvertObserverOuputLambdaKey)

	_ = g.AddBranch(ConvertObserverOuputLambdaKey, compose.NewGraphBranch(observerDecisionBranch, map[string]bool{
		GuidelineProposerPromptTplKey: true,
		DoriaPromptTplKey:             true,
	}))

	_ = g.AddEdge(DoriaPromptTplKey, DoriaChatModelKey)
	_ = g.AddEdge(DoriaChatModelKey, compose.END)

	return g, nil
}

func saveInputToState(ctx context.Context, input map[string]any, state *state) (map[string]any, error) {
	if p, ok := input["prompt"].(string); ok {
		state.prompt = p
	}
	if h, ok := input["history"].([]*schema.Message); ok {
		state.history = h
	}
	if g, ok := input["guidelines"].([]*Guideline); ok {
		state.guidelines = g
		state.guidelinesString = FormatGuidelines(g)
		input["guidelines"] = state.guidelinesString
	}

	return input, nil
}

func activeGuidelinesLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	var (
		history    []*schema.Message
		prompt     string
		guidelines []*Guideline
	)

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		history = state.history
		prompt = state.prompt
		guidelines = state.guidelines
		return nil
	}); err != nil {
		return nil, err
	}

	var payload InputPayload

	if err := json.Unmarshal([]byte(input.Content), &payload); err != nil {
		return nil, err
	}

	guidelineEvaluations := payload.GuidelineEvaluations

	activeGuidelines := make([]*Guideline, 0)
	if len(guidelineEvaluations) == 0 {
		return map[string]any{
			"history":           history,
			"prompt":            prompt,
			"active_guidelines": activeGuidelines,
			"has_tool":          false,
		}, nil
	}

	maxScore := math.MinInt32
	for _, ge := range guidelineEvaluations {
		if ge.ConditionApplies && ge.AppliesScore > maxScore {
			maxScore = ge.AppliesScore
		}
	}

	if maxScore == math.MinInt32 {
		return map[string]any{
			"history":           history,
			"prompt":            prompt,
			"active_guidelines": activeGuidelines,
			"has_tool":          false,
		}, nil
	}

	guidelineMap := make(map[string]*Guideline)
	for _, g := range guidelines {
		guidelineMap[g.ID] = g
	}

	for _, ge := range guidelineEvaluations {
		if ge.ConditionApplies && ge.AppliesScore == maxScore {
			if guideline, ok := guidelineMap[ge.GuidelineID]; ok {
				activeGuidelines = append(activeGuidelines, guideline)
			}
		}
	}

	toolsInfo := ""
	hasTool := false

	for _, g := range activeGuidelines {
		if len(g.Tools) > 0 {
			hasTool = true
		}

		info, err := FormatToolsInfo(ctx, g.Tools)
		if err != nil {
			return nil, err
		}
		if info != "" {
			toolsInfo += info + "\n"
		}
	}

	activeGuidelinesString := FormatGuidelines(activeGuidelines)

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		state.activeGuidelines = activeGuidelines
		state.activeGuidelinesString = activeGuidelinesString
		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]any{
		"history":           history,
		"prompt":            prompt,
		"active_guidelines": activeGuidelinesString,
		"tools_info":        toolsInfo,
		"has_tool":          hasTool,
	}, nil
}

func toolCallingLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	var (
		history                []*schema.Message
		prompt                 string
		activeGuidelines       []*Guideline
		activeGuidelinesString string
	)

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		activeGuidelines = state.activeGuidelines
		activeGuidelinesString = state.activeGuidelinesString
		prompt = state.prompt
		history = state.history
		return nil
	}); err != nil {
		return nil, err
	}

	var evaluation EvaluationResponse
	if err := json.Unmarshal([]byte(input.Content), &evaluation); err != nil {
		return nil, fmt.Errorf("解析评估结果失败: %w, 原始响应: %s", err, input.Content)
	}

	toolEvaluation := FindBestTool(&evaluation)

	tools := make([]tool.BaseTool, 0)
	for _, g := range activeGuidelines {
		tools = append(tools, g.Tools...)
	}

	toolsOutput, err := ExecuteTool(ctx, tools, toolEvaluation)
	if err != nil {
		return nil, err
	}

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		state.toolOutput = toolsOutput
		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]any{
		"tools_output":      toolsOutput,
		"history":           history,
		"prompt":            prompt,
		"active_guidelines": activeGuidelinesString,
	}, nil
}

func convertObserverOutputLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	var (
		history                []*schema.Message
		prompt                 string
		activeGuidelinesString string
		guidelinesString       string
		toolsOutput            string
	)

	if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
		activeGuidelinesString = state.activeGuidelinesString
		guidelinesString = state.guidelinesString
		prompt = state.prompt
		history = state.history
		toolsOutput = state.toolOutput
		return nil
	}); err != nil {
		return nil, err
	}

	observerOutput := ObserverOutput{}
	if err := json.Unmarshal([]byte(input.Content), &observerOutput); err != nil {
		return nil, err
	}

	return map[string]any{
		"history":           history,
		"prompt":            prompt,
		"active_guidelines": activeGuidelinesString,
		"guidelines":        guidelinesString,
		"toward":            observerOutput.Toward,
		"tools_output":      toolsOutput,
	}, nil
}

func toolCallingDecisionBranch(ctx context.Context, input map[string]any) (endNode string, err error) {
	hasTool, ok := input["has_tool"].(bool)

	if !ok || !hasTool {
		input["tools_output"] = ""

		epoch := 0
		if err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
			state.epoch = state.epoch + 1
			epoch = state.epoch
			return nil
		}); err != nil {
			return compose.END, err
		}

		if epoch >= 2 {
			return DoriaPromptTplKey, nil
		}

		return DoriaPromptTplKey, nil
	}

	return ToolCallerPromptTplKey, nil
}

func observerDecisionBranch(ctx context.Context, input map[string]any) (endNode string, err error) {
	toward, ok := input["toward"].(bool)
	if !ok || !toward {
		input["tools_output"] = fmt.Sprintf("为了达成用户的要求，曾经调用了工具，工具的输出结果为：\n%s\n但是并不能解决用户的问题，你需要重新进行对用户的问题进行评估\n", input["tools_output"])
		return GuidelineProposerPromptTplKey, nil
	}
	return DoriaPromptTplKey, nil
}
