package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Fl0rencess720/Doria/src/common/rag"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type RagTool struct {
	hr *rag.HybridRetriever
}

type RAGToolInput struct {
	Query string `json:"query" jsonschema:"description=查询文本,required"`
}

func NewRAGTool(ctx context.Context) (tool.InvokableTool, error) {
	hr, err := rag.NewHybridRetriever(ctx)
	if err != nil {
		return nil, err
	}
	if hr == nil {
		return nil, fmt.Errorf("hybrid retriever is nil")
	}

	return &RagTool{
		hr: hr,
	}, nil
}

func (t *RagTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "RAG",
		Desc: "基于检索的自动生成工具",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "查询文本",
				Required: true,
			},
		}),
	}, nil
}

func (w *RagTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input RAGToolInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", err
	}

	if input.Query == "" {
		return "", fmt.Errorf("query is empty")
	}

	hr := w.hr
	docs, err := hr.Retrieve(ctx, input.Query)
	if err != nil {
		return "检索出错: " + err.Error(), err
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

	return result, nil
}
