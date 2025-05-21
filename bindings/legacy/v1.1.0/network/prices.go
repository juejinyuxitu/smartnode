package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the block number which network prices are current for
func GetPricesBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketNetworkPricesAddress *common.Address) (uint64, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, opts)
	if err != nil {
		return 0, err
	}
	pricesBlock := new(*big.Int)
	if err := rocketNetworkPrices.Call(opts, pricesBlock, "getPricesBlock"); err != nil {
		return 0, fmt.Errorf("error getting network prices block: %w", err)
	}
	return (*pricesBlock).Uint64(), nil
}

// Get the current network RPL price in ETH
func GetRPLPrice(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketNetworkPricesAddress *common.Address) (*big.Int, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, opts)
	if err != nil {
		return nil, err
	}
	rplPrice := new(*big.Int)
	if err := rocketNetworkPrices.Call(opts, rplPrice, "getRPLPrice"); err != nil {
		return nil, fmt.Errorf("error getting network RPL price: %w", err)
	}
	return *rplPrice, nil
}

// Estimate the gas of SubmitPrices
func EstimateSubmitPricesGas(rp *rocketpool.RocketPool, block uint64, rplPrice *big.Int, effectiveRplStake *big.Int, opts *bind.TransactOpts, legacyRocketNetworkPricesAddress *common.Address) (rocketpool.GasInfo, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkPrices.GetTransactionGasInfo(opts, "submitPrices", big.NewInt(int64(block)), rplPrice, effectiveRplStake)
}

// Submit network prices and total effective RPL stake for an epoch
func SubmitPrices(rp *rocketpool.RocketPool, block uint64, rplPrice, effectiveRplStake *big.Int, opts *bind.TransactOpts, legacyRocketNetworkPricesAddress *common.Address) (common.Hash, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkPrices.Transact(opts, "submitPrices", big.NewInt(int64(block)), rplPrice, effectiveRplStake)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting network prices: %w", err)
	}
	return tx.Hash(), nil
}

// Check if the network is currently in consensus about the RPL price, or if it is still reaching consensus
func InConsensus(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketNetworkPricesAddress *common.Address) (bool, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, opts)
	if err != nil {
		return false, err
	}
	isInConsensus := new(bool)
	if err := rocketNetworkPrices.Call(opts, isInConsensus, "inConsensus"); err != nil {
		return false, fmt.Errorf("error getting consensus status: %w", err)
	}
	return *isInConsensus, nil
}

// Returns the latest block number that oracles should be reporting prices for
func GetLatestReportablePricesBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketNetworkPricesAddress *common.Address) (*big.Int, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, legacyRocketNetworkPricesAddress, opts)
	if err != nil {
		return nil, err
	}
	latestReportableBlock := new(*big.Int)
	if err := rocketNetworkPrices.Call(opts, latestReportableBlock, "getLatestReportableBlock"); err != nil {
		return nil, fmt.Errorf("error getting latest reportable block: %w", err)
	}
	return *latestReportableBlock, nil
}

// Get contracts
var rocketNetworkPricesLock sync.Mutex

func getRocketNetworkPrices(rp *rocketpool.RocketPool, address *common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkPricesLock.Lock()
	defer rocketNetworkPricesLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_1_0.GetContract("rocketNetworkPrices", opts)
	} else {
		return rp.VersionManager.V1_1_0.GetContractWithAddress("rocketNetworkPrices", *address)
	}
}
