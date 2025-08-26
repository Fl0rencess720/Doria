package configs

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(GetServiceName)

func GetServiceName() string {
	return viper.GetString("server.grpc.name")
}
