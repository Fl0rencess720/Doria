package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageRepo, NewUserRepo, NewTTSRepo,
	NewMateRepo, NewSignalingRepo, NewImageClient, NewUserClient, NewTTSClient, NewMateClient)
