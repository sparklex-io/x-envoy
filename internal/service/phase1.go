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
	"time"
)

// RunPhase1Services listens for Local Exit Root (LER) update events of different chains
// 1. Get chains to monitor from config
// 2. Build clients for these chains
// 3. Query the last Local Exit Root(LER) from source chains
// 4. Query the corresponding LER from SparkleX
// 5. If the LER from source chain is different from SparkleX, update SparkleX
func RunPhase1Services(config config.Config) {
	var wg conc.WaitGroup
	for _, service := range config.Services {
		wg.Go(func() { RunPhase1Service(config, service) })
	}
	wg.Wait()
}

func RunPhase1Service(config config.Config, service config.ServiceConfig) {
	// FIXME: viper transfers map key to lowercase
	fromConfig, ok := config.Networks[service.From]
	if !ok {
		log.Error().Msgf("Network %s not found in config", service.From)
		panic("network not found in config")
	}

	// FIXME: Private key can not starts with '0x'
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

	fromClient, err := client.NewEvmClient(fromConfig.URL, fromConfig.GEREAddress, fromConfig.BridgeAddress, config.PrivateKey)
	if err != nil {
		log.Error().Msgf("Failed to create source chain client: %v", err)
		panic(err)
	}
	networkID, err := fromClient.GetNetworkID()
	if err != nil {
		log.Error().Msgf("Failed to get network ID from source chain: %v", err)
		panic(err)
	}

	for {
		sourceLER, err := fromClient.GetLastLocalExitRoot()
		if err != nil {
			log.Error().Msgf("Failed to get last LER from source chain: %v", err)
			panic(err)
		}
		depositCount, err := fromClient.GetDepositCount()
		if err != nil {
			log.Error().Msgf("Failed to get deposit count from source chain: %v", err)
			panic(err)
		}

		correspondingLER, err := sClient.GetNetworkLER(networkID)
		if err != nil {
			log.Error().Msgf("Failed to get corresponding LER from SparkleX: %v", err)
			panic(err)
		}

		equal := bytes.Equal(sourceLER[:], correspondingLER[:])
		log.Info().
			Str("direction", fmt.Sprintf("%v->sparklex", service.From)).
			Uint64("count", depositCount.Uint64()).
			Bool("isSynced", equal).
			Send()
		if !bytes.Equal(sourceLER[:], correspondingLER[:]) {
			tx, err := sClient.UpdateLocalExitRoot(networkID, uint32(depositCount.Uint64()), sourceLER)
			if err != nil {
				log.Error().Msgf("Failed to update LER in SparkleX: %v", err)
				panic(err)
			}
			log.Info().
				Str("direction", fmt.Sprintf("%v->sparklex", service.From)).
				Msgf("Update LER in SparkleX: %v", tx.Hash().Hex())
			_, err = bind.WaitMined(context.TODO(), sClient.Client(), tx)
			if err != nil {
				log.Error().
					Str("direction", fmt.Sprintf("%v->sparklex", service.From)).
					Msgf("Failed to wait for LER update tx to be mined: %v", err)
				panic(err)
			}
			continue
		}
		time.Sleep(20 * time.Second)
	}
}
