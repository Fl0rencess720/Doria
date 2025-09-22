package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

var toolsRegistry = make(map[string]tool.InvokableTool)

func GetRAGTool() (tool.InvokableTool, error) {
	if t, ok := toolsRegistry["RAG"]; ok {
		return t, nil
	}

	return nil, fmt.Errorf("RAG tool not found")
}

func Init() error {
	ctx := context.Background()

	t, err := NewRAGTool(ctx)
	if err != nil {
		return err
	}
	toolsRegistry["RAG"] = t

	return nil
}
