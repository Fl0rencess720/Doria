package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/cloudwego/eino/components/tool"
)

type ArgumentEvaluation struct {
	IsAvailable bool        `json:"is_available"`
	Value       interface{} `json:"value"`
	Rationale   string      `json:"rationale"`
}

type ToolEvaluation struct {
	ToolName               string                        `json:"tool_name"`
	SubtletiesToBeAwareOf  string                        `json:"subtleties_to_be_aware_of"`
	ApplicabilityRationale string                        `json:"applicability_rationale"`
	ApplicabilityScore     int                           `json:"applicability_score"`
	ArgumentEvaluations    map[string]ArgumentEvaluation `json:"argument_evaluations"`
	ShouldRun              bool                          `json:"should_run"`
}

type EvaluationResponse struct {
	ToolEvaluations []ToolEvaluation `json:"tool_evaluations"`
}

func FormatToolsInfo(ctx context.Context, tools []tool.BaseTool) (string, error) {
	toolDescriptions := make([]string, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return "", err
		}

		paramInfo := "无参数"
		if info.ParamsOneOf != nil {
			if jsonSchema, err := info.ParamsOneOf.ToJSONSchema(); err == nil && jsonSchema != nil && jsonSchema.Properties != nil {
				var paramDescs []string
				for pair := jsonSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
					paramName := pair.Key
					paramSchema := pair.Value

					required := "可选"
					for _, req := range jsonSchema.Required {
						if req == paramName {
							required = "必需"
							break
						}
					}

					paramDescs = append(paramDescs, fmt.Sprintf("- %s (%s, %s): %s",
						paramName, paramSchema.Type, required, paramSchema.Description))
				}
				if len(paramDescs) > 0 {
					paramInfo = strings.Join(paramDescs, "\n")
				}
			}
		}

		toolDesc := fmt.Sprintf(`工具名称: %s  
		描述: %s  
		参数定义:  
		%s`, info.Name, info.Desc, paramInfo)

		toolDescriptions = append(toolDescriptions, toolDesc)
	}

	return strings.Join(toolDescriptions, "\n\n"), nil
}

func FindBestTool(evaluation *EvaluationResponse) *ToolEvaluation {
	var bestTool *ToolEvaluation
	maxScore := math.MinInt

	for i := range evaluation.ToolEvaluations {
		tool := &evaluation.ToolEvaluations[i]
		if tool.ShouldRun && tool.ApplicabilityScore > maxScore {
			maxScore = tool.ApplicabilityScore
			bestTool = tool
		}
	}

	return bestTool
}

func ExecuteTool(ctx context.Context, tools []tool.BaseTool, toolEval *ToolEvaluation) (string, error) {
	var targetTool tool.BaseTool
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		if info.Name == toolEval.ToolName {
			targetTool = t
			break
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("未找到工具: %s", toolEval.ToolName)
	}

	params := make(map[string]interface{})
	for paramName, argEval := range toolEval.ArgumentEvaluations {
		if argEval.IsAvailable && argEval.Value != nil {
			params[paramName] = argEval.Value
		}
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("参数序列化失败: %w", err)
	}

	if invokable, ok := targetTool.(tool.InvokableTool); ok {
		return invokable.InvokableRun(ctx, string(paramsJSON))
	}

	return "", fmt.Errorf("工具 %s 不支持调用", toolEval.ToolName)
}
