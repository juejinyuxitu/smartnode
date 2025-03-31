package pdao

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func proposeSecurityCouncilReplace(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the list of members
	membersResponse, err := rp.SecurityMembers()
	if err != nil {
		return fmt.Errorf("error getting list of security council members: %w", err)
	}

	// Get the address of the member to replace
	var oldID string
	var oldAddress common.Address
	oldAddressString := c.String("existing-address")
	if oldAddressString == "" {
		options := make([]string, len(membersResponse.Members))
		for i, member := range membersResponse.Members {
			options[i] = fmt.Sprintf("%d: %s (%s), joined %s\n", i+1, member.ID, member.Address, time.Unix(int64(member.JoinedTime), 0))
		}
		selection, _ := prompt.Select("Which member would you like to replace?", options)
		member := membersResponse.Members[selection]
		oldID = member.ID
		oldAddress = member.Address
	} else {
		oldAddress, err = cliutils.ValidateAddress("address", oldAddressString)
		if err != nil {
			return err
		}
		found := false
		for _, member := range membersResponse.Members {
			if member.Address == oldAddress {
				oldID = member.ID
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("address %s is not a member of the security council", oldAddress.Hex())
		}
	}

	// Get the new ID
	newID := c.String("new-id")
	if newID == "" {
		newID = prompt.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	newID, err = cliutils.ValidateDAOMemberID("id", newID)
	if err != nil {
		return err
	}

	// Get the new address
	newAddressString := c.String("new-address")
	if newAddressString == "" {
		newAddressString = prompt.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	newAddress, err := cliutils.ValidateAddress("address", newAddressString)
	if err != nil {
		return err
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeReplaceMemberOfSecurityCouncil(oldAddress, newID, newAddress)
	if err != nil {
		return err
	}
	if !canResponse.CanPropose {
		fmt.Println("Cannot propose security council replacement:")
		if canResponse.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to propose removing %s (%s) from the security council and inviting %s (%s)?", oldID, oldAddress.Hex(), newID, newAddress.Hex()))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeReplaceMemberOfSecurityCouncil(oldAddress, newID, newAddress, canResponse.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Proposing replace in security council...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
