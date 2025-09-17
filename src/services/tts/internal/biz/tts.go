package biz

type TTSUseCase struct {
	repo TTSRepo
}

type TTSRepo interface {
}

func NewTTSUseCase(repo TTSRepo) *TTSUseCase {
	return &TTSUseCase{repo: repo}
}
