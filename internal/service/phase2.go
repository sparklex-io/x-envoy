package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sourcegraph/conc"
	"github.com/sparklex-io/envoy/config"
	"github.com/sparklex-io/envoy/internal/client"
	"github.com/sparklex-io/envoy/internal/log"
	"github.com/status-im/keycard-go/hexutils"
	"time"
)

func RunPhase2Services(config config.Config) {
	var wg conc.WaitGroup
	for _, service := range config.Services {
		wg.Go(func() { RunPhase2Service(config, service) })
	}
	wg.Wait()
}

func RunPhase2Service(config config.Config, service config.ServiceConfig) {
	toConfig, ok := config.Networks[service.To]
	if !ok {
		log.Error().Msgf("Network %s not found in config", service.From)
		panic("network not found in config")
	}

	sClient, err := client.NewSparkleXClient(
		config.SparkleX.URL,
		config.SparkleX.NetworkManagerAddress,
		config.SparkleX.ReducerAddress,
		config.SparkleX.GERManagerAddress,
		config.PrivateKey)
	if err != nil {
		log.Error().Msgf("Failed to create SparkleX client: %v", err)
		panic(err)
	}

	toClient, err := client.NewEvmClient(toConfig.URL, toConfig.GEREAddress, toConfig.BridgeAddress, config.PrivateKey)
	if err != nil {
		log.Error().Msgf("Failed to create source chain client: %v", err)
		panic(err)
	}

	for {
		lastGEREvent, err := sClient.GetLastUpdateGEREvent()
		if err != nil {
			log.Error().Msgf("Failed to get last GER from SparkleX: %v", err)
			panic(err)
		}
		if lastGEREvent == nil {
			log.Info().Msg("No GER found in SparkleX")
			time.Sleep(13 * time.Second)
			continue
		}

		relayBlockNum, err := toClient.QueryGER(lastGEREvent.NewRoot)
		if err != nil {
			log.Error().Msgf("Failed to query GER in target chain: %v", err)
			panic(err)
		}
		if relayBlockNum.Uint64() != 0 {
			log.Info().
				Str("direction", fmt.Sprintf("sparklex->%v", service.To)).
				Bool("isSynced", true).
				Send()
			time.Sleep(13 * time.Second)
			continue
		}
		lastRelayedGER, err := toClient.GetLastUpdateGEREvent()
		if err != nil {
			log.Error().Msgf("Failed to get last GER from target chain: %v", err)
			panic(err)
		}
		if lastRelayedGER != nil && bytes.Equal(lastRelayedGER.NewRoot[:], lastGEREvent.NewRoot[:]) {
			log.Info().
				Str("direction", fmt.Sprintf("sparklex->%v", service.To)).
				Msg("GER in target chain is up-to-date")
			time.Sleep(13 * time.Second)
			continue
		}
		if lastRelayedGER != nil {
			log.Info().
				Str("direction", fmt.Sprintf("sparklex->%v", service.To)).
				Msgf("Last GER in target chain: %s", hexutils.BytesToHex(lastRelayedGER.NewRoot[:]))
		}

		tx, err := toClient.UpdateGER(lastGEREvent.NewRoot)
		if err != nil {
			log.Error().Msgf("Failed to update GER in target chain: %v", err)
			panic(err)
		}
		log.Info().
			Str("direction", fmt.Sprintf("sparklex->%v", service.To)).
			Msgf("Update GER in target chain: %s", tx.Hash().Hex())
		_, err = bind.WaitMined(context.TODO(), toClient.Client(), tx)
		if err != nil {
			log.Error().Msgf("Failed to wait for GER update tx to be mined: %v", err)
			panic(err)
		}
	}
}
