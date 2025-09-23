package biz

type MemoryRepo interface {
}

type MemoryUseCase struct {
	repo MemoryRepo
}

func NewMemoryUseCase(repo MemoryRepo) *MemoryUseCase {
	return &MemoryUseCase{
		repo: repo,
	}
}

func (uc *MemoryUseCase) ProcessMemory() {

}

func (uc *MemoryUseCase) RetrieveMemory() {

}
