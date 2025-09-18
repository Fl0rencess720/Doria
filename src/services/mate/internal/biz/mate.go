package biz

type MateUseCase struct {
	repo MateRepo
}

type MateRepo interface {
}

func NewMateUseCase(repo MateRepo) *MateUseCase {
	return &MateUseCase{repo: repo}
}
