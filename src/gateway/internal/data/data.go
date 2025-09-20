package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageRepo, NewChatRepo, NewUserRepo, NewTTSRepo,
	NewMateRepo, NewImageClient, NewChatClient, NewUserClient, NewTTSClient, NewMateClient)
