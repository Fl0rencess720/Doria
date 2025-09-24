package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	SegmentOverviewTemplateKey  = "segment_overview_tpl"
	SegmentOverviewChatModelKey = "segment_overview_chat_model"
)

func buildSegmentOverviewGraph(_ context.Context, cm model.ToolCallingChatModel) (*compose.Graph[map[string]any, *schema.Message], error) {
	g := compose.NewGraph[map[string]any, *schema.Message]()
	segmentOverviewTpl := newSegmentOverviewTemplate()

	g.AddChatTemplateNode(SegmentOverviewTemplateKey, segmentOverviewTpl)

	g.AddChatModelNode(SegmentOverviewChatModelKey, cm)

	g.AddEdge(compose.START, SegmentOverviewTemplateKey)
	g.AddEdge(SegmentOverviewTemplateKey, SegmentOverviewChatModelKey)
	g.AddEdge(SegmentOverviewChatModelKey, compose.END)

	return g, nil
}
