package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func initializeVotingWithDelegate(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the address
	delegateAddressString := c.String("address")
	if delegateAddressString == "" {
		delegateAddressString = prompt.Prompt("Please enter the delegate's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	delegateAddress, err := cliutils.ValidateAddress("delegateAddress", delegateAddressString)
	if err != nil {
		return err
	}

	resp, err := rp.CanInitializeVotingWithDelegate(delegateAddress)
	if err != nil {
		return fmt.Errorf("error calling get-voting-initialized: %w", err)
	}

	if resp.VotingInitialized {
		fmt.Println("Node voting was already initialized")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to initialize voting?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Initialize voting
	response, err := rp.InitializeVotingWithDelegate(delegateAddress)
	if err != nil {
		return fmt.Errorf("error calling initialize-voting: %w", err)
	}

	fmt.Printf("Initializing voting...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error initializing voting: %w", err)
	}

	// Log & return
	fmt.Println("Successfully initialized voting.")
	return nil
}
