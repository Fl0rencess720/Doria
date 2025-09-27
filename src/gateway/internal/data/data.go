package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageRepo, NewUserRepo, NewTTSRepo,
	NewMateRepo, NewImageClient, NewUserClient, NewTTSClient, NewMateClient)
