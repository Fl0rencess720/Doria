package service

import (
	"context"

	memoryapi "github.com/Fl0rencess720/Doria/src/rpc/memory"
	"github.com/Fl0rencess720/Doria/src/services/memory/internal/biz"
)

func (s *MemoryService) GetMemory(ctx context.Context, req *memoryapi.GetMemoryRequest) (*memoryapi.GetMemoryResponse, error) {
	memory, err := s.memoryUseCase.RetrieveMemory(ctx, uint(req.UserId), req.Prompt)
	if err != nil {
		return nil, err
	}

	stm := make([]*memoryapi.ShortMidTermMemory, len(memory))
	mtm := make([]*memoryapi.ShortMidTermMemory, len(memory))
	ltm := make([]*memoryapi.LongTermMemory, len(memory))

	for _, m := range memory {
		switch m.Status {
		case biz.QAStatusInSTM:
			stm = append(stm, &memoryapi.ShortMidTermMemory{
				UserInput:   m.UserInput,
				AgentOutput: m.AgentOutput,
			})
		case biz.QAStatusInMTM:
			mtm = append(mtm, &memoryapi.ShortMidTermMemory{
				UserInput:   m.UserInput,
				AgentOutput: m.AgentOutput,
			})
		case biz.QAStatusInLTM:
			ltm = append(ltm, &memoryapi.LongTermMemory{
				Context: m.UserInput + m.AgentOutput,
			})
		default:
			continue
		}
	}

	resp := &memoryapi.GetMemoryResponse{
		ShortTermMemory: stm,
		MidTermMemory:   mtm,
		LongTermMemory:  ltm,
	}

	return resp, nil
}
