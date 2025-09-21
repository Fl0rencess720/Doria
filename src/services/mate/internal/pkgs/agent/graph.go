package agent

import (
	"context"
	"encoding/json"
	"math"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	GuidelineProposerPromptTplKey = "guideline_proposer_prompt"
	ToolCallerPromptTplKey        = "tool_caller_prompt"
	DoriaPromptTplKey             = "doria_prompt"
	GuidelineProposerChatModelKey = "guideline_proposer_chat_model"
	ToolCallerChatModelKey        = "tool_caller_chat_model"
	DoriaChatModelKey             = "doria_chat_model"

	ActiveGuidelinesLambdaKey = "active_guidelines_lambda"
	ToolCallingLambdaKey      = "tool_calling_lambda"
)

type state struct {
	history    []*schema.Message
	prompt     string
	hr         *rag.HybridRetriever
	guidelines string
}

func buildChatGraph(_ context.Context, mainCM model.ToolCallingChatModel,
	jsonCM model.ToolCallingChatModel, hr *rag.HybridRetriever) (*compose.Graph[map[string]any, *schema.Message], error) {
	compose.RegisterSerializableType[state]("state")

	guidelineProposerTpl := newGuidelineProposerResponseTemplate()
	toolCallerTpl := newToolCallerResponseTemplate()
	doriaTpl := newDoriaResponseTemplate()

	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *state {
			return &state{
				hr: hr,
			}
		}))

	_ = g.AddChatTemplateNode(GuidelineProposerPromptTplKey, guidelineProposerTpl, compose.WithStatePreHandler(saveInputToState))
	_ = g.AddChatTemplateNode(ToolCallerPromptTplKey, toolCallerTpl)
	_ = g.AddChatTemplateNode(DoriaPromptTplKey, doriaTpl)

	_ = g.AddChatModelNode(GuidelineProposerChatModelKey, jsonCM)
	_ = g.AddChatModelNode(ToolCallerChatModelKey, jsonCM)
	_ = g.AddChatModelNode(DoriaChatModelKey, mainCM)

	_ = g.AddLambdaNode(ActiveGuidelinesLambdaKey, compose.InvokableLambda(activeGuidelinesLambda))
	_ = g.AddLambdaNode(ToolCallingLambdaKey, compose.InvokableLambda(toolCallingLambda))

	_ = g.AddEdge(compose.START, GuidelineProposerPromptTplKey)
	_ = g.AddEdge(GuidelineProposerPromptTplKey, GuidelineProposerChatModelKey)
	_ = g.AddEdge(GuidelineProposerChatModelKey, ActiveGuidelinesLambdaKey)

	_ = g.AddBranch(ActiveGuidelinesLambdaKey, compose.NewGraphBranch(toolCallingDecisionBranch, map[string]bool{
		ToolCallerPromptTplKey: true,
		DoriaPromptTplKey:      true,
	}))

	_ = g.AddEdge(ToolCallerPromptTplKey, ToolCallerChatModelKey)
	_ = g.AddEdge(ToolCallerChatModelKey, ToolCallingLambdaKey)
	_ = g.AddEdge(ToolCallingLambdaKey, DoriaPromptTplKey)
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
	if g, ok := input["guidelines"].(string); ok {
		state.guidelines = g
	}

	return input, nil
}

type InputPayload struct {
	GuidelineEvaluations []*GuidelineEvaluation `json:"guideline_evaluations"`
}

func activeGuidelinesLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	var (
		history    []*schema.Message
		prompt     string
		guidelines string
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
			"history":    history,
			"prompt":     prompt,
			"guidelines": activeGuidelines,
			"has_tool":   false,
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
			"history":    history,
			"prompt":     prompt,
			"guidelines": activeGuidelines,
			"has_tool":   false,
		}, nil
	}

	guidelinesList := make([]*Guideline, 0)
	if err := json.Unmarshal([]byte(guidelines), &guidelinesList); err != nil {
		return nil, err
	}

	guidelineMap := make(map[string]*Guideline)
	for _, g := range guidelinesList {
		guidelineMap[g.ID] = g
	}

	for _, ge := range guidelineEvaluations {
		if ge.ConditionApplies && ge.AppliesScore == maxScore {
			if guideline, ok := guidelineMap[ge.GuidelineID]; ok {
				activeGuidelines = append(activeGuidelines, guideline)
			}
		}
	}

	activeGuidelinesBytes, err := json.Marshal(activeGuidelines)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"history":    history,
		"prompt":     prompt,
		"guidelines": string(activeGuidelinesBytes),
		"has_tool":   false,
	}, nil
}

func toolCallingLambda(ctx context.Context, input *schema.Message) (map[string]any, error) {
	return map[string]any{}, nil
}

func toolCallingDecisionBranch(ctx context.Context, input map[string]any) (endNode string, err error) {
	hasTool, ok := input["has_tool"].(bool)
	if !ok || !hasTool {
		input["tools_output"] = ""
		return DoriaPromptTplKey, nil
	}

	return ToolCallerPromptTplKey, nil
}
