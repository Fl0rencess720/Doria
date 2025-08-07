package profiling

import (
	"os"
	"runtime"

	"github.com/grafana/pyroscope-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitPyroscope(appName string) {
	if viper.GetString("pyroscope.state") != "enable" {
		zap.L().Info("User close Pyroscope, the service would not run.",
			zap.String("appName", appName))
		return
	}

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: appName,
		ServerAddress:   viper.GetString("PYROSCOPE_ADDRESS"),
		Logger:          zap.L().Sugar(),
		Tags:            map[string]string{"hostname": os.Getenv("HOSTNAME")},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	if err != nil {
		zap.L().Error("Error while run Pyroscope.",
			zap.String("appName", appName),
			zap.Error(err))
		return
	}
}
