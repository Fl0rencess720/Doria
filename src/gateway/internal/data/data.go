package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageRepo, NewChatRepo, NewUserRepo, NewTTSRepo,
	NewImageClient, NewChatClient, NewUserClient, NewTTSClient)
