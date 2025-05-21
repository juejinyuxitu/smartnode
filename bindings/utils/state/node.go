package state

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	legacyNodeBatchSize  int = 100
	nodeAddressBatchSize int = 1000
)

// Complete details for a node
type NativeNodeDetails struct {
	Exists                           bool           `json:"exists"`
	RegistrationTime                 *big.Int       `json:"registration_time"`
	TimezoneLocation                 string         `json:"timezone_location"`
	FeeDistributorInitialised        bool           `json:"fee_distributor_initialised"`
	FeeDistributorAddress            common.Address `json:"fee_distributor_address"`
	RewardNetwork                    *big.Int       `json:"reward_network"`
	RplStake                         *big.Int       `json:"rpl_stake"`
	EffectiveRPLStake                *big.Int       `json:"effective_rpl_stake"`
	MinimumRPLStake                  *big.Int       `json:"minimum_rpl_stake"`
	MaximumRPLStake                  *big.Int       `json:"maximum_rpl_stake"`
	EthMatched                       *big.Int       `json:"eth_matched"`
	EthMatchedLimit                  *big.Int       `json:"eth_matched_limit"`
	MinipoolCount                    *big.Int       `json:"minipool_count"`
	BalanceETH                       *big.Int       `json:"balance_eth"`
	BalanceRETH                      *big.Int       `json:"balance_reth"`
	BalanceRPL                       *big.Int       `json:"balance_rpl"`
	BalanceOldRPL                    *big.Int       `json:"balance_old_rpl"`
	DepositCreditBalance             *big.Int       `json:"deposit_credit_balance"`
	DistributorBalanceUserETH        *big.Int       `json:"distributor_balance_user_eth"` // Must call CalculateAverageFeeAndDistributorShares to get this
	DistributorBalanceNodeETH        *big.Int       `json:"distributor_balance_node_eth"` // Must call CalculateAverageFeeAndDistributorShares to get this
	WithdrawalAddress                common.Address `json:"withdrawal_address"`
	PendingWithdrawalAddress         common.Address `json:"pending_withdrawal_address"`
	SmoothingPoolRegistrationState   bool           `json:"smoothing_pool_registration_state"`
	SmoothingPoolRegistrationChanged *big.Int       `json:"smoothing_pool_registration_changed"`
	NodeAddress                      common.Address `json:"node_address"`
	AverageNodeFee                   *big.Int       `json:"average_node_fee"` // Must call CalculateAverageFeeAndDistributorShares to get this
	CollateralisationRatio           *big.Int       `json:"collateralisation_ratio"`
	DistributorBalance               *big.Int       `json:"distributor_balance"`
}

func timeMax(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func timeMin(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// Returns whether the node is eligible for bonuses, and the start and end times of its eligibility
func (nnd *NativeNodeDetails) IsEligibleForBonuses(eligibleStart time.Time, eligibleEnd time.Time) (bool, time.Time, time.Time) {
	// Nodes are not eligible for bonuses if they never opted into the smoothing pool
	registeredTime := time.Unix(nnd.SmoothingPoolRegistrationChanged.Int64(), 0)
	if registeredTime.Unix() == 0 {
		return false, time.Time{}, time.Time{}
	}

	// Nodes are eligible for bonuses if they were in the Smoothing Pool for a portion of the interval
	if nnd.SmoothingPoolRegistrationState {
		return registeredTime.Before(eligibleEnd), timeMax(registeredTime, eligibleStart), eligibleEnd
	}

	// Nodes that weren't opted in at the end of the interval are eligible if they opted out during the interval
	return registeredTime.Before(eligibleEnd), timeMax(registeredTime, eligibleStart), timeMin(registeredTime, eligibleEnd)
}

// Gets the details for a node using the efficient multicall contract
func GetNativeNodeDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, nodeAddress common.Address) (NativeNodeDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}
	details := NativeNodeDetails{
		NodeAddress:               nodeAddress,
		AverageNodeFee:            big.NewInt(0),
		CollateralisationRatio:    big.NewInt(0),
		DistributorBalanceUserETH: big.NewInt(0),
		DistributorBalanceNodeETH: big.NewInt(0),
	}

	addNodeDetailsCalls(contracts, contracts.Multicaller, &details, nodeAddress)

	_, err := contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return NativeNodeDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	// Get the node's ETH balance
	details.BalanceETH, err = rp.Client.BalanceAt(context.Background(), nodeAddress, opts.BlockNumber)
	if err != nil {
		return NativeNodeDetails{}, err
	}

	// Get the distributor balance
	distributorBalance, err := rp.Client.BalanceAt(context.Background(), details.FeeDistributorAddress, opts.BlockNumber)
	if err != nil {
		return NativeNodeDetails{}, err
	}

	// Do some postprocessing on the node data
	details.DistributorBalance = distributorBalance

	// Fix the effective stake
	if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
		details.EffectiveRPLStake.SetUint64(0)
	}

	return details, nil
}

// Gets the details for all nodes using the efficient multicall contract
func GetAllNativeNodeDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]NativeNodeDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	count := len(addresses)
	nodeDetails := make([]NativeNodeDetails, count)

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	for i := 0; i < count; i += legacyNodeBatchSize {
		i := i
		max := i + legacyNodeBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				address := addresses[j]
				details := &nodeDetails[j]
				details.NodeAddress = address
				details.AverageNodeFee = big.NewInt(0)
				details.DistributorBalanceUserETH = big.NewInt(0)
				details.DistributorBalanceNodeETH = big.NewInt(0)
				details.CollateralisationRatio = big.NewInt(0)

				addNodeDetailsCalls(contracts, mc, details, address)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node details: %w", err)
	}

	// Get the balances of the nodes
	distributorAddresses := make([]common.Address, count)
	balances, err := contracts.BalanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node balances: %w", err)
	}
	for i, details := range nodeDetails {
		nodeDetails[i].BalanceETH = balances[i]
		distributorAddresses[i] = details.FeeDistributorAddress
	}

	// Get the balances of the distributors
	balances, err = contracts.BalanceBatcher.GetEthBalances(distributorAddresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting distributor balances: %w", err)
	}

	// Do some postprocessing on the node data
	for i := range nodeDetails {
		details := &nodeDetails[i]
		details.DistributorBalance = balances[i]

		// Fix the effective stake
		if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
			details.EffectiveRPLStake.SetUint64(0)
		}
	}

	return nodeDetails, nil
}

func (node *NativeNodeDetails) WasOptedInAt(t time.Time) bool {
	if node.SmoothingPoolRegistrationState {
		// If a node is opted in, check if the check time is after the opt-in time
		return t.After(time.Unix(node.SmoothingPoolRegistrationChanged.Int64(), 0))
	}

	// If the node isn't opted in and was never opted in, it's not opted in
	if node.SmoothingPoolRegistrationChanged.Cmp(big.NewInt(0)) == 0 {
		return false
	}

	// If a node is opted out, but was opted in, check if the check time is before the opt-out time
	return t.Before(time.Unix(node.SmoothingPoolRegistrationChanged.Int64(), 0))
}

// Calculate the average node fee and user/node shares of the distributor's balance
func (node *NativeNodeDetails) CalculateAverageFeeAndDistributorShares(minipoolDetails []*NativeMinipoolDetails) error {

	// Calculate the total of all fees for staking minipools that aren't finalized
	totalFee := big.NewInt(0)
	eligibleMinipools := int64(0)
	for _, mpd := range minipoolDetails {
		if mpd.Status == types.Staking && !mpd.Finalised {
			totalFee.Add(totalFee, mpd.NodeFee)
			eligibleMinipools++
		}
	}

	// Get the average fee (0 if there aren't any minipools)
	if eligibleMinipools > 0 {
		node.AverageNodeFee.Div(totalFee, big.NewInt(eligibleMinipools))
	}

	// Get the user and node portions of the distributor balance
	distributorBalance := big.NewInt(0).Set(node.DistributorBalance)
	if distributorBalance.Cmp(big.NewInt(0)) > 0 {
		nodeBalance := big.NewInt(0)
		nodeBalance.Mul(distributorBalance, big.NewInt(1e18))
		nodeBalance.Div(nodeBalance, node.CollateralisationRatio)

		userBalance := big.NewInt(0)
		userBalance.Sub(distributorBalance, nodeBalance)

		if eligibleMinipools == 0 {
			// Split it based solely on the collateralisation ratio if there are no minipools (and hence no average fee)
			node.DistributorBalanceNodeETH = big.NewInt(0).Set(nodeBalance)
			node.DistributorBalanceUserETH = big.NewInt(0).Sub(distributorBalance, nodeBalance)
		} else {
			// Amount of ETH given to the NO as a commission
			commissionEth := big.NewInt(0)
			commissionEth.Mul(userBalance, node.AverageNodeFee)
			commissionEth.Div(commissionEth, big.NewInt(1e18))

			node.DistributorBalanceNodeETH.Add(nodeBalance, commissionEth)                         // Node gets their portion + commission on user portion
			node.DistributorBalanceUserETH.Sub(distributorBalance, node.DistributorBalanceNodeETH) // User gets balance - node share
		}

	} else {
		// No distributor balance
		node.DistributorBalanceNodeETH = big.NewInt(0)
		node.DistributorBalanceUserETH = big.NewInt(0)
	}

	return nil
}

// Get all node addresses using the multicaller
func getNodeAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, opts *bind.CallOpts) ([]common.Address, error) {
	// Get minipool count
	nodeCount, err := node.GetNodeCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, nodeCount)

	// Run the getters in batches
	count := int(nodeCount)
	for i := 0; i < count; i += nodeAddressBatchSize {
		i := i
		max := i + nodeAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketNodeManager, &addresses[j], "getNodeAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	return addresses, nil
}

// Add all of the calls for the node details to the multicaller
func addNodeDetailsCalls(contracts *NetworkContracts, mc *multicall.MultiCaller, details *NativeNodeDetails, address common.Address) {
	mc.AddCall(contracts.RocketNodeManager, &details.Exists, "getNodeExists", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RegistrationTime, "getNodeRegistrationTime", address)
	mc.AddCall(contracts.RocketNodeManager, &details.TimezoneLocation, "getNodeTimezoneLocation", address)
	mc.AddCall(contracts.RocketNodeManager, &details.FeeDistributorInitialised, "getFeeDistributorInitialised", address)
	mc.AddCall(contracts.RocketNodeDistributorFactory, &details.FeeDistributorAddress, "getProxyAddress", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RewardNetwork, "getRewardNetwork", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.RplStake, "getNodeRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EffectiveRPLStake, "getNodeEffectiveRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.MinimumRPLStake, "getNodeMinimumRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.MaximumRPLStake, "getNodeMaximumRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EthMatched, "getNodeETHMatched", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EthMatchedLimit, "getNodeETHMatchedLimit", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.MinipoolCount, "getNodeMinipoolCount", address)
	mc.AddCall(contracts.RocketTokenRETH, &details.BalanceRETH, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPL, &details.BalanceRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPLFixedSupply, &details.BalanceOldRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketStorage, &details.WithdrawalAddress, "getNodeWithdrawalAddress", address)
	mc.AddCall(contracts.RocketStorage, &details.PendingWithdrawalAddress, "getNodePendingWithdrawalAddress", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationState, "getSmoothingPoolRegistrationState", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationChanged, "getSmoothingPoolRegistrationChanged", address)

	// Atlas
	mc.AddCall(contracts.RocketNodeDeposit, &details.DepositCreditBalance, "getNodeDepositCredit", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.CollateralisationRatio, "getNodeETHCollateralisationRatio", address)
}
