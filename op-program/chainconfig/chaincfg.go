package chainconfig

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/params"
)

var enabledFromBedrockBlock = uint64(0)

var OPGoerliChainConfig, OPSepoliaChainConfig, OPMainnetChainConfig *params.ChainConfig

func init() {
	mustLoadConfig := func(chainID uint64) *params.ChainConfig {
		cfg, err := params.LoadOPStackChainConfig(chainID)
		if err != nil {
			panic(err)
		}
		return cfg
	}
	OPGoerliChainConfig = mustLoadConfig(420)
	OPSepoliaChainConfig = mustLoadConfig(11155420)
	OPSepoliaChainConfig = mustLoadConfig(10)
}

var L2ChainConfigsByName = map[string]*params.ChainConfig{
	"goerli":  OPGoerliChainConfig,
	"sepolia": OPSepoliaChainConfig,
	"mainnet": OPMainnetChainConfig,
}

func RollupConfigByChainID(chainID uint64) (*rollup.Config, error) {
	network := chaincfg.L2ChainIDToNetworkName[strconv.FormatUint(chainID, 10)]
	if network == "" {
		return nil, fmt.Errorf("unknown chain ID: %d", chainID)
	}
	config, ok := chaincfg.NetworksByName[network]
	if !ok {
		return nil, fmt.Errorf("unknown network %s for chain ID %d", network, chainID)
	}
	return &config, nil
}

func ChainConfigByChainID(chainID uint64) (*params.ChainConfig, error) {
	network := chaincfg.L2ChainIDToNetworkName[strconv.FormatUint(chainID, 10)]
	chainConfig, ok := L2ChainConfigsByName[network]
	if !ok {
		return nil, fmt.Errorf("unknown network %s for chain ID %d", network, chainID)
	}
	return chainConfig, nil
}
