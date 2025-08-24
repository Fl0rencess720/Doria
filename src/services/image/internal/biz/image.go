package biz

type ImageRepo interface {
}

type ImageUseCase struct {
	imageRepo ImageRepo
}

func NewImageUseCase(imageRepo ImageRepo) *ImageUseCase {
	return &ImageUseCase{
		imageRepo: imageRepo,
	}
}

func (*ImageUseCase) GenerateTextOfImage(imageData []byte, style string) (string, error) {
	return "", nil
}
