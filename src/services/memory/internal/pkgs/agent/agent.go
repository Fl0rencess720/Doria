package agent

import (
	"context"
	"strings"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/pkgs/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	runnable compose.Runnable[map[string]any, *schema.Message]
}

func NewAgent(ctx context.Context) (*Agent, error) {
	cm, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}

	sg, err := buildSegmentOverviewGraph(ctx, cm)
	if err != nil {
		return nil, err
	}

	runnable, err := sg.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &Agent{
		runnable: runnable,
	}, nil
}

func (a *Agent) GenSegmentOverview(ctx context.Context, qas []*models.Page) (string, error) {
	var builder strings.Builder

	for _, qa := range qas {
		qaPair := utils.BuildQAPair(qa.UserInput, qa.AgentOutput)
		builder.WriteString(qaPair)
	}

	qasString := builder.String()

	response, err := a.runnable.Invoke(ctx, map[string]any{
		"qas": qasString,
	})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
