package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sparklex-io/envoy/generated/bridge"
	"github.com/sparklex-io/envoy/generated/gere"
	"math/big"
	"sort"
)

type EvmClient struct {
	client      *ethclient.Client
	gereManager *gere.GlobalExitRootExternal
	bridge      *bridge.Bridge
	privateKey  *ecdsa.PrivateKey
}

func NewEvmClient(url, gereAddress, bridgeAddress, key string) (*EvmClient, error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(gereAddress) {
		return nil, errors.New("invalid hex address for GERE")
	}
	gereManager, err := gere.NewGlobalExitRootExternal(common.HexToAddress(gereAddress), client)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(bridgeAddress) {
		return nil, errors.New("invalid hex address for bridge")
	}
	bridgeManager, err := bridge.NewBridge(common.HexToAddress(bridgeAddress), client)
	if err != nil {
		return nil, err
	}
	return &EvmClient{
		client:      client,
		gereManager: gereManager,
		bridge:      bridgeManager,
		privateKey:  privateKey,
	}, nil
}

func (c *EvmClient) Client() *ethclient.Client {
	return c.client
}

func (c *EvmClient) GetLastLocalExitRoot() ([32]byte, error) {
	resp, err := c.gereManager.LastLocalExitRoot(nil)
	if err != nil {
		return [32]byte{}, err
	}

	return resp, nil
}

func (c *EvmClient) GetNetworkID() (uint32, error) {
	networkID, err := c.bridge.NetworkID(nil)
	if err != nil {
		return 0, err
	}
	return networkID, nil
}

func (c *EvmClient) GetDepositCount() (*big.Int, error) {
	depositCount, err := c.bridge.DepositCount(nil)
	if err != nil {
		return nil, err
	}
	return depositCount, nil
}

func (c *EvmClient) GetLastUpdateGEREvent() (*gere.GlobalExitRootExternalUpdateGlobalExitRoot, error) {
	blockNumber, err := c.client.BlockNumber(context.TODO())
	if err != nil {
		return nil, err
	}

	it, err := c.gereManager.FilterUpdateGlobalExitRoot(&bind.FilterOpts{
		Start: blockNumber - 256,
		End:   nil,
	})
	if err != nil {
		return nil, err
	}

	var logs []*gere.GlobalExitRootExternalUpdateGlobalExitRoot
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

func (c *EvmClient) UpdateGER(newRoot [32]byte) (*types.Transaction, error) {
	chainID, err := c.client.ChainID(context.TODO())
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, chainID)
	if err != nil {
		return nil, err
	}

	tx, err := c.gereManager.UpdateGlobalExitRoot(auth, newRoot)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *EvmClient) QueryGER(root [32]byte) (*big.Int, error) {
	return c.gereManager.GlobalExitRootMap(nil, root)
}
