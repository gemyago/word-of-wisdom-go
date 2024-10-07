package config

import (
	"fmt"
	"word-of-wisdom-go/internal/di"

	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type configValueProvider struct {
	cfg        *viper.Viper
	configPath string
	diPath     string
}

func provideConfigValue(cfg *viper.Viper, path string) configValueProvider {
	if !cfg.IsSet(path) {
		panic(fmt.Errorf("config key not found: %s", path))
	}
	return configValueProvider{cfg, path, "config." + path}
}

func (p configValueProvider) asInt() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetInt(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asInt64() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetInt64(p.configPath), dig.Name(p.diPath))
}

/*
func (p configValueProvider) asString() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetString(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asBool() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetBool(p.configPath), dig.Name(p.diPath))
}
*/

func (p configValueProvider) asDuration() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetDuration(p.configPath), dig.Name(p.diPath))
}

func Provide(container *dig.Container, cfg *viper.Viper) error {
	return di.ProvideAll(container,
		// tcp server config
		provideConfigValue(cfg, "tcpServer.port").asInt(),
		provideConfigValue(cfg, "tcpServer.maxSessionDuration").asDuration(),

		// client config
		provideConfigValue(cfg, "client.ioTimeout").asDuration(),
		provideConfigValue(cfg, "client.maxSessionDuration").asDuration(),

		// monitoring config
		provideConfigValue(cfg, "monitoring.windowDuration").asDuration(),
		provideConfigValue(cfg, "monitoring.maxUnverifiedClientRequests").asInt64(),
		provideConfigValue(cfg, "monitoring.maxUnverifiedRequests").asInt64(),

		// challenges config
		provideConfigValue(cfg, "challenges.maxSolveChallengeDuration").asDuration(),
	)
}
