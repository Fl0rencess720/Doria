package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageUsecase, NewChatUseCase, NewUserUsecase,
	NewTTSUsecase, NewMateUsecase)
