package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sparklex-io/envoy/generated/ger"
	"github.com/sparklex-io/envoy/generated/network_manager"
	"github.com/sparklex-io/envoy/generated/reducer"
	"github.com/sparklex-io/envoy/internal/mapper"
	"github.com/status-im/keycard-go/hexutils"
	"sort"
)

type SparkleXClient struct {
	client         *ethclient.Client
	networkManager *network_manager.NetworkManager
	gerManager     *ger.GlobalExitRoot
	reducer        *reducer.Reducer
	privateKey     *ecdsa.PrivateKey
	mapper         mapper.Mapper
	p              float64
}

// NewSparkleXClient creates a new SparkleX client
// Parameters:
// - url: the URL of the SparkleX node
// - nwAddress: the address of the network manager contract
// Returns:
// - *SparkleXClient: the SparkleX client
func NewSparkleXClient(url, nwAddress, reducerAddress, gerAddress, key string) (*SparkleXClient, error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(nwAddress) {
		return nil, errors.New("invalid network manager address")
	}
	addr := common.HexToAddress(nwAddress)

	if !common.IsHexAddress(reducerAddress) {
		return nil, errors.New("invalid network manager address")
	}
	rAddr := common.HexToAddress(reducerAddress)

	if !common.IsHexAddress(gerAddress) {
		return nil, errors.New("invalid network manager address")
	}
	gAddr := common.HexToAddress(gerAddress)

	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	networkManager, err := network_manager.NewNetworkManager(addr, client)
	if err != nil {
		return nil, err
	}
	reducerContract, err := reducer.NewReducer(rAddr, client)
	if err != nil {
		return nil, err
	}
	gerManager, err := ger.NewGlobalExitRoot(gAddr, client)
	if err != nil {
		return nil, err
	}

	return &SparkleXClient{
		client:         client,
		networkManager: networkManager,
		reducer:        reducerContract,
		gerManager:     gerManager,
		privateKey:     privateKey,
		mapper:         mapper.Mapper{},
		p:              mapper.Tau / mapper.K,
	}, nil
}

func (c *SparkleXClient) GetNetworkLER(networkID uint32) ([32]byte, error) {
	resp, err := c.networkManager.NetworkIDToNetworkData(nil, networkID)
	if err != nil {
		return [32]byte{}, err
	}

	return resp.LastLocalExitRoot, nil
}

func (c *SparkleXClient) UpdateLocalExitRoot(networkID, depositCount uint32, newLocalExitRoot [32]byte) (*types.Transaction, error) {
	chainID, err := c.client.ChainID(context.TODO())
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, chainID)
	if err != nil {
		return nil, err
	}

	// TODO: Stake value should be queried from the SparkleX.
	networkIDBytes := fmt.Sprintf("%08x", networkID)
	message := hexutils.HexToBytes(networkIDBytes)
	message = append(message, newLocalExitRoot[:]...)
	payload, err := c.mapper.BuildVotePayload(c.privateKey, c.p, message, 1_000_000)
	if err != nil {
		return nil, err
	}
	payload.Message.NetworkId = networkID
	payload.Message.DepositCount = depositCount
	payload.Message.State = newLocalExitRoot
	tx, err := c.reducer.VoteFastForMessageState(auth, payload)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *SparkleXClient) GetLastUpdateGEREvent() (*ger.GlobalExitRootUpdateGlobalExitRoot, error) {
	blockNumber, err := c.client.BlockNumber(context.TODO())
	if err != nil {
		return nil, err
	}

	it, err := c.gerManager.FilterUpdateGlobalExitRoot(&bind.FilterOpts{
		Start: blockNumber - 1000,
		End:   nil,
	})
	if err != nil {
		return nil, err
	}

	var logs []*ger.GlobalExitRootUpdateGlobalExitRoot
	for it.Next() {
		logs = append(logs, it.Event)
	}
	if len(logs) == 0 {
		return nil, nil
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Raw.BlockNumber < logs[j].Raw.BlockNumber
	})
	return logs[len(logs)-1], nil
}

func (c *SparkleXClient) Client() *ethclient.Client {
	return c.client
}
