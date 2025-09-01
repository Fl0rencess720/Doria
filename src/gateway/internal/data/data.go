package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewImageRepo, NewChatRepo, NewUserRepo, NewImageClient, NewChatClient, NewUserClient)
