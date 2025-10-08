package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageUsecase, NewUserUsecase,
	NewTTSUsecase, NewMateUsecase, NewSignalingUsecase)
