package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
)

// Estimate the gas of ProposeSetMulti
func EstimateProposeSetMultiGas(rp *rocketpool.RocketPool, message string, contractNames []string, settingPaths []string, settingTypes []types.ProposalSettingType, values []any, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	encodedValues, err := abiEncodeMultiValues(settingTypes, values)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error ABI encoding values: %w", err)
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingMulti", contractNames, settingPaths, settingTypes, encodedValues)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error setting multi-set proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update multiple Protocol DAO settings at once
func ProposeSetMulti(rp *rocketpool.RocketPool, message string, contractNames []string, settingPaths []string, settingTypes []types.ProposalSettingType, values []any, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	encodedValues, err := abiEncodeMultiValues(settingTypes, values)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error ABI encoding values: %w", err)
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingMulti", contractNames, settingPaths, settingTypes, encodedValues)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error setting multi-set proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetBool
func EstimateProposeSetBoolGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error setting bool setting proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a bool Protocol DAO setting
func ProposeSetBool(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error setting bool setting proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetUint
func EstimateProposeSetUintGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a uint Protocol DAO setting
func ProposeSetUint(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetAddress
func EstimateProposeSetAddressGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingAddress", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set address setting proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update an address Protocol DAO setting
func ProposeSetAddress(rp *rocketpool.RocketPool, message, contractName, settingPath string, value common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingAddress", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set address setting proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetRewardsPercentage
func EstimateProposeSetRewardsPercentageGas(rp *rocketpool.RocketPool, message string, odaoPercentage *big.Int, pdaoPercentage *big.Int, nodePercentage *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingRewardsClaimers", odaoPercentage, pdaoPercentage, nodePercentage)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set rewards-claimers percent proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update the allocations of RPL rewards
func ProposeSetRewardsPercentage(rp *rocketpool.RocketPool, message string, odaoPercentage *big.Int, pdaoPercentage *big.Int, nodePercentage *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingRewardsClaimers", odaoPercentage, pdaoPercentage, nodePercentage)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set rewards-claimers percent proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeOneTimeTreasurySpend
func EstimateProposeOneTimeTreasurySpendGas(rp *rocketpool.RocketPool, message, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryOneTimeSpend", invoiceID, recipient, amount)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set spend-treasury percent proposal payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to spend a portion of the Rocket Pool treasury one time
func ProposeOneTimeTreasurySpend(rp *rocketpool.RocketPool, message, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryOneTimeSpend", invoiceID, recipient, amount)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set spend-treasury percent proposal payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeRecurringTreasurySpend
func EstimateProposeRecurringTreasurySpendGas(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryNewContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(startTime.Unix()), big.NewInt(int64(numberOfPeriods)))
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalTreasuryNewContract payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to spend a portion of the Rocket Pool treasury in a recurring manner
func ProposeRecurringTreasurySpend(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryNewContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(startTime.Unix()), big.NewInt(int64(numberOfPeriods)))
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalTreasuryNewContract payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeRecurringTreasurySpendUpdate
func EstimateProposeRecurringTreasurySpendUpdateGas(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryUpdateContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(int64(numberOfPeriods)))
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalTreasuryUpdateContract payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a recurrint Rocket Pool treasury spending plan
func ProposeRecurringTreasurySpendUpdate(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryUpdateContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(int64(numberOfPeriods)))
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalTreasuryUpdateContract payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeInviteToSecurityCouncil
func EstimateProposeInviteToSecurityCouncilGas(rp *rocketpool.RocketPool, message string, id string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityInvite", id, address)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityInvite payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to invite a member to the security council
func ProposeInviteToSecurityCouncil(rp *rocketpool.RocketPool, message string, id string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityInvite", id, address)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityInvite payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeKickFromSecurityCouncil
func EstimateProposeKickFromSecurityCouncilGas(rp *rocketpool.RocketPool, message string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKick", address)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityKick payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to kick a member from the security council
func ProposeKickFromSecurityCouncil(rp *rocketpool.RocketPool, message string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKick", address)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityKick payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeKickMultiFromSecurityCouncil
func EstimateProposeKickMultiFromSecurityCouncilGas(rp *rocketpool.RocketPool, message string, addresses []common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKickMulti", addresses)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityKickMulti payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to kick multiple members from the security council
func ProposeKickMultiFromSecurityCouncil(rp *rocketpool.RocketPool, message string, addresses []common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKickMulti", addresses)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityKickMulti payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeReplaceSecurityCouncilMember
func EstimateProposeReplaceSecurityCouncilMemberGas(rp *rocketpool.RocketPool, message string, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityReplace", existingMemberAddress, newMemberID, newMemberAddress)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityReplace payload: %w", err)
	}
	return estimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to replace a member of the security council with another one in a single TX
func ProposeReplaceSecurityCouncilMember(rp *rocketpool.RocketPool, message string, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityReplace", existingMemberAddress, newMemberID, newMemberAddress)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityReplace payload: %w", err)
	}
	return submitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Get the ABI encoding of multiple values for a ProposeSettingMulti call
func abiEncodeMultiValues(settingTypes []types.ProposalSettingType, values []any) ([][]byte, error) {
	// Sanity check the lengths
	settingCount := len(settingTypes)
	if settingCount != len(values) {
		return nil, fmt.Errorf("settingTypes and values must be the same length")
	}
	if settingCount == 0 {
		return [][]byte{}, nil
	}

	// ABI encode each value
	results := make([][]byte, settingCount)
	for i, settingType := range settingTypes {
		var encodedArg []byte
		switch settingType {
		case types.ProposalSettingType_Uint256:
			arg, success := values[i].(*big.Int)
			if !success {
				return nil, fmt.Errorf("value %d is not a *big.Int, but the setting type is Uint256", i)
			}
			encodedArg = math.U256Bytes(big.NewInt(0).Set(arg))

		case types.ProposalSettingType_Bool:
			arg, success := values[i].(bool)
			if !success {
				return nil, fmt.Errorf("value %d is not a bool, but the setting type is Bool", i)
			}
			if arg {
				encodedArg = math.PaddedBigBytes(common.Big1, 32)
			} else {
				encodedArg = math.PaddedBigBytes(common.Big0, 32)
			}

		case types.ProposalSettingType_Address:
			arg, success := values[i].(common.Address)
			if !success {
				return nil, fmt.Errorf("value %d is not an address, but the setting type is Address", i)
			}
			encodedArg = common.LeftPadBytes(arg.Bytes(), 32)

		default:
			return nil, fmt.Errorf("unknown proposal setting type [%v]", settingType)
		}
		results[i] = encodedArg
	}

	return results, nil
}

// Get contracts
var rocketDAOProtocolProposalsLock sync.Mutex

func getRocketDAOProtocolProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolProposalsLock.Lock()
	defer rocketDAOProtocolProposalsLock.Unlock()
	return rp.GetContract("rocketDAOProtocolProposals", opts)
}
