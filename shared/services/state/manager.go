package state

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type NetworkStateManager struct {
	rp  *rocketpool.RocketPool
	bc  beacon.Client
	log *log.ColorLogger

	// Memoized Beacon config
	beaconConfig *beacon.Eth2Config

	// Multicaller and batch balance contract addresses
	multicaller    common.Address
	balanceBatcher common.Address
}

// Create a new manager for the network state
func NewNetworkStateManager(
	rp *rocketpool.RocketPool,
	contracts config.StateManagerContracts,
	bc beacon.Client,
	log *log.ColorLogger,
) *NetworkStateManager {

	// Create the manager
	return &NetworkStateManager{
		rp:             rp,
		bc:             bc,
		log:            log,
		multicaller:    contracts.Multicaller,
		balanceBatcher: contracts.BalanceBatcher,
	}
}

func (m *NetworkStateManager) getBeaconConfig() (*beacon.Eth2Config, error) {
	if m.beaconConfig != nil {
		return m.beaconConfig, nil
	}

	// Get the Beacon config info
	beaconConfig, err := m.bc.GetEth2Config()
	if err != nil {
		return nil, err
	}
	m.beaconConfig = &beaconConfig

	return m.beaconConfig, nil
}

// Get the state of the network using the latest Execution layer block
func (m *NetworkStateManager) GetHeadState() (*NetworkState, error) {
	targetSlot, err := m.getHeadSlot()
	if err != nil {
		return nil, fmt.Errorf("error getting latest Beacon slot: %w", err)
	}
	return m.createNetworkState(targetSlot)
}

// Get the state of the network for a single node using the latest Execution layer block, along with the total effective RPL stake for the network
func (m *NetworkStateManager) GetHeadStateForNode(nodeAddress common.Address) (*NetworkState, error) {
	targetSlot, err := m.getHeadSlot()
	if err != nil {
		return nil, fmt.Errorf("error getting latest Beacon slot: %w", err)
	}
	return m.createNetworkStateForNode(targetSlot, nodeAddress)
}

// Get the state of the network at the provided Beacon slot
func (m *NetworkStateManager) GetStateForSlot(slotNumber uint64) (*NetworkState, error) {
	return m.createNetworkState(slotNumber)
}

// Gets the latest valid block
func (m *NetworkStateManager) GetLatestBeaconBlock() (beacon.BeaconBlock, error) {
	targetSlot, err := m.getHeadSlot()
	if err != nil {
		return beacon.BeaconBlock{}, fmt.Errorf("error getting head slot: %w", err)
	}
	return m.getLatestProposedBeaconBlock(targetSlot)
}

// Gets the latest valid finalized block
func (m *NetworkStateManager) GetLatestFinalizedBeaconBlock() (beacon.BeaconBlock, error) {
	beaconConfig, err := m.getBeaconConfig()
	if err != nil {
		return beacon.BeaconBlock{}, fmt.Errorf("error getting Beacon config: %w", err)
	}
	head, err := m.bc.GetBeaconHead()
	if err != nil {
		return beacon.BeaconBlock{}, fmt.Errorf("error getting Beacon chain head: %w", err)
	}
	targetSlot := head.FinalizedEpoch*beaconConfig.SlotsPerEpoch + (beaconConfig.SlotsPerEpoch - 1)
	return m.getLatestProposedBeaconBlock(targetSlot)
}

// Gets the Beacon slot for the latest execution layer block
func (m *NetworkStateManager) getHeadSlot() (uint64, error) {
	beaconConfig, err := m.getBeaconConfig()
	if err != nil {
		return 0, fmt.Errorf("error getting Beacon config: %w", err)
	}
	// Get the latest EL block
	latestBlockHeader, err := m.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest EL block: %w", err)
	}

	// Get the corresponding Beacon slot based on the timestamp
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)
	secondsSinceGenesis := uint64(latestBlockTime.Sub(genesisTime).Seconds())
	targetSlot := secondsSinceGenesis / beaconConfig.SecondsPerSlot
	return targetSlot, nil
}

// Gets the target Beacon block, or if it was missing, the first one under it that wasn't missing
func (m *NetworkStateManager) getLatestProposedBeaconBlock(targetSlot uint64) (beacon.BeaconBlock, error) {
	for {
		// Try to get the current block
		block, exists, err := m.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
		if err != nil {
			return beacon.BeaconBlock{}, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			m.logLine("Slot %d was missing, trying the previous one...", targetSlot)
			targetSlot--
		} else {
			return block, nil
		}
	}
}

// Logs a line if the logger is specified
func (m *NetworkStateManager) logLine(format string, v ...interface{}) {
	if m.log != nil {
		m.log.Printlnf(format, v...)
	}
}
