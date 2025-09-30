package agent

import (
	"context"
	"strings"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

type agent struct {
	segmentOverviewRunnable     compose.Runnable[map[string]any, *schema.Message]
	knowledgeExtractionRunnable compose.Runnable[map[string]any, *schema.Message]
}

func NewAgent() biz.LLMAgent {
	ctx := context.Background()

	cm, err := newChatModel(ctx)
	if err != nil {
		zap.L().Panic("New Chat Model error", zap.Error(err))
	}

	sg, err := buildSegmentOverviewGraph(ctx, cm)
	if err != nil {
		zap.L().Panic("New Segment Overview Graph error", zap.Error(err))
	}

	kg, err := buildKnowledgeExtractionGraph(ctx, cm)
	if err != nil {
		zap.L().Panic("New Knowledge Extraction Graph error", zap.Error(err))
	}

	sr, err := sg.Compile(ctx)
	if err != nil {
		zap.L().Panic("New Segment Overview Runnable error", zap.Error(err))
	}

	kr, err := kg.Compile(ctx)
	if err != nil {
		zap.L().Panic("New Knowledge Extraction Runnable error", zap.Error(err))
	}

	return &agent{
		segmentOverviewRunnable:     sr,
		knowledgeExtractionRunnable: kr,
	}
}

func (a *agent) GenSegmentOverview(ctx context.Context, qas []*models.Page) (string, error) {
	var builder strings.Builder

	for _, qa := range qas {
		qaPair := utils.BuildQAPair(qa.UserInput, qa.AgentOutput)
		builder.WriteString(qaPair)
	}

	qasString := builder.String()

	response, err := a.segmentOverviewRunnable.Invoke(ctx, map[string]any{
		"qas": qasString,
	})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func (a *agent) GenKnowledgeExtraction(ctx context.Context, qas []*models.Page, knowledge string) (string, error) {
	var builder strings.Builder

	for _, qa := range qas {
		qaPair := utils.BuildQAPair(qa.UserInput, qa.AgentOutput)
		builder.WriteString(qaPair)
	}

	qasString := builder.String()

	response, err := a.knowledgeExtractionRunnable.Invoke(ctx, map[string]any{
		"qas":       qasString,
		"knowledge": knowledge,
	})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
