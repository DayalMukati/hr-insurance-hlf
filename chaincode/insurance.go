package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract for insurance claim processing
type SmartContract struct {
	contractapi.Contract
}

// Policyholder represents an insured individual
type Policyholder struct {
	PolicyID string `json:"policyID"`
	Name     string `json:"name"`
	Balance  int    `json:"balance"`
}

// InsuranceClaim represents an insurance claim
type InsuranceClaim struct {
	ClaimID   string `json:"claimID"`
	PolicyID  string `json:"policyID"`
	Amount    int    `json:"amount"`
	Reason    string `json:"reason"`
	Status    string `json:"status"` // Pending, Approved, Rejected
}

// RegisterPolicyholder registers a new policyholder
func (s *SmartContract) RegisterPolicyholder(ctx contractapi.TransactionContextInterface, policyID string, name string, balance int) error {
	exists, err := ctx.GetStub().GetState(policyID)
	if err != nil {
		return fmt.Errorf("failed to check policyholder existence: %v", err)
	}
	if len(exists) > 0 {
		return fmt.Errorf("policyholder already exists")
	}

	policyholder := Policyholder{
		PolicyID: policyID,
		Name:     name,
		Balance:  balance,
	}

	policyholderJSON, err := json.Marshal(policyholder)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(policyID, policyholderJSON)
}

// FileClaim allows a policyholder to file an insurance claim
func (s *SmartContract) FileClaim(ctx contractapi.TransactionContextInterface, claimID string, policyID string, amount int, reason string) error {
	// Check if policyholder exists
	policyholderJSON, err := ctx.GetStub().GetState(policyID)
	if err != nil {
		return fmt.Errorf("failed to read policyholder state: %v", err)
	}
	if policyholderJSON == nil {
		return fmt.Errorf("policyholder does not exist")
	}

	claim := InsuranceClaim{
		ClaimID:  claimID,
		PolicyID: policyID,
		Amount:   amount,
		Reason:   reason,
		Status:   "Pending",
	}

	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(claimID, claimJSON)
}

// ReviewClaim allows an insurer to approve or reject a claim
func (s *SmartContract) ReviewClaim(ctx contractapi.TransactionContextInterface, claimID string, approve bool) error {
	claimJSON, err := ctx.GetStub().GetState(claimID)
	if err != nil {
		return fmt.Errorf("failed to read claim state: %v", err)
	}
	if claimJSON == nil {
		return fmt.Errorf("claim does not exist")
	}

	var claim InsuranceClaim
	err = json.Unmarshal(claimJSON, &claim)
	if err != nil {
		return err
	}

	if claim.Status != "Pending" {
		return fmt.Errorf("claim is already processed")
	}

	// Approve or Reject the claim
	if approve {
		claim.Status = "Approved"

		// Update policyholder's balance
		policyholderJSON, err := ctx.GetStub().GetState(claim.PolicyID)
		if err != nil {
			return fmt.Errorf("failed to read policyholder state: %v", err)
		}
		if policyholderJSON == nil {
			return fmt.Errorf("policyholder does not exist")
		}

		var policyholder Policyholder
		err = json.Unmarshal(policyholderJSON, &policyholder)
		if err != nil {
			return err
		}

		policyholder.Balance += claim.Amount

		updatedPolicyholderJSON, err := json.Marshal(policyholder)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(policyholder.PolicyID, updatedPolicyholderJSON)
		if err != nil {
			return err
		}

	} else {
		claim.Status = "Rejected"
	}

	updatedClaimJSON, err := json.Marshal(claim)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(claimID, updatedClaimJSON)
}

// GetClaimDetails retrieves an insurance claim
func (s *SmartContract) GetClaimDetails(ctx contractapi.TransactionContextInterface, claimID string) (*InsuranceClaim, error) {
	claimJSON, err := ctx.GetStub().GetState(claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to read claim state: %v", err)
	}
	if claimJSON == nil {
		return nil, fmt.Errorf("claim does not exist")
	}

	var claim InsuranceClaim
	err = json.Unmarshal(claimJSON, &claim)
	if err != nil {
		return nil, err
	}

	return &claim, nil
}

// GetPolicyholderDetails retrieves a policyholder's details
func (s *SmartContract) GetPolicyholderDetails(ctx contractapi.TransactionContextInterface, policyID string) (*Policyholder, error) {
	policyholderJSON, err := ctx.GetStub().GetState(policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to read policyholder state: %v", err)
	}
	if policyholderJSON == nil {
		return nil, fmt.Errorf("policyholder does not exist")
	}

	var policyholder Policyholder
	err = json.Unmarshal(policyholderJSON, &policyholder)
	if err != nil {
		return nil, err
	}

	return &policyholder, nil
}

// Main function to start the chaincode
func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating insurance claim chaincode: %s", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting insurance claim chaincode: %s", err)
	}
}
