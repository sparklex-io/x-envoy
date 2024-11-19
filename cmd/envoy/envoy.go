package main

import (
	"github.com/sourcegraph/conc"
	config "github.com/sparklex-io/envoy/config"
	"github.com/sparklex-io/envoy/internal/log"
	"github.com/sparklex-io/envoy/internal/service"
)

func main() {
	userConfig, err := config.LoadConfig("config", "$HOME/.envoy/", ".")
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		panic("failed to load config")
	}

	var wg conc.WaitGroup

	// FIXME: panic recovery
	// Writes GERs to SparkleX
	wg.Go(func() { service.RunPhase1Services(userConfig) })
	wg.Go(func() { service.RunPhase2Services(userConfig) })
	wg.Wait()
}
