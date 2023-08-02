package fault

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-challenger/fault/types"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// faultResponder implements the [Responder] interface to send onchain transactions.
type faultResponder struct {
	log log.Logger

	txMgr txmgr.TxManager

	fdgAddr common.Address
	fdgAbi  *abi.ABI

	preimageOracleAddr common.Address
	preimageOracleAbi  *abi.ABI
}

// NewFaultResponder returns a new [faultResponder].
func NewFaultResponder(logger log.Logger, txManagr txmgr.TxManager, fdgAddr common.Address, preimageOracleAddr common.Address) (*faultResponder, error) {
	fdgAbi, err := bindings.FaultDisputeGameMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	preimageOracleAbi, err := bindings.PreimageOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return &faultResponder{
		log:                logger,
		txMgr:              txManagr,
		fdgAddr:            fdgAddr,
		fdgAbi:             fdgAbi,
		preimageOracleAddr: preimageOracleAddr,
		preimageOracleAbi:  preimageOracleAbi,
	}, nil
}

// buildFaultDefendData creates the transaction data for the Defend function.
func (r *faultResponder) buildFaultDefendData(parentContractIndex int, pivot [32]byte) ([]byte, error) {
	return r.fdgAbi.Pack(
		"defend",
		big.NewInt(int64(parentContractIndex)),
		pivot,
	)
}

// buildFaultAttackData creates the transaction data for the Attack function.
func (r *faultResponder) buildFaultAttackData(parentContractIndex int, pivot [32]byte) ([]byte, error) {
	return r.fdgAbi.Pack(
		"attack",
		big.NewInt(int64(parentContractIndex)),
		pivot,
	)
}

// buildResolveData creates the transaction data for the Resolve function.
func (r *faultResponder) buildResolveData() ([]byte, error) {
	return r.fdgAbi.Pack("resolve")
}

// BuildTx builds the transaction for the [faultResponder].
func (r *faultResponder) BuildTx(ctx context.Context, response types.Claim) ([]byte, error) {
	if response.DefendsParent() {
		txData, err := r.buildFaultDefendData(response.ParentContractIndex, response.ValueBytes())
		if err != nil {
			return nil, err
		}
		return txData, nil
	} else {
		txData, err := r.buildFaultAttackData(response.ParentContractIndex, response.ValueBytes())
		if err != nil {
			return nil, err
		}
		return txData, nil
	}
}

// PopulateOracleData uploads the preimage oracle data into the onchain preimage oracle contract.
func (r *faultResponder) PopulateOracleData(ctx context.Context, data types.PreimageOracleData) error {
	var txData []byte
	var err error
	if data.IsLocal {
		txData, err = r.buildLocalOracleData(data)
		if err != nil {
			return fmt.Errorf("local oracle tx data build: %w", err)
		}
	} else {
		txData, err = r.buildGlobalOracleData(data)
		if err != nil {
			return fmt.Errorf("global oracle tx data build: %w", err)
		}
	}
	return r.sendTxAndWait(ctx, txData)
}

// buildLocalOracleData takes the local preimage key and data
// and creates tx data to load the key, data pair into the
// PreimageOracle contract from the FaultDisputeGame contract call.
//
// Encoded call to: addLocalData(uint256 _ident, uint256 _partOffset) external
func (r *faultResponder) buildLocalOracleData(data types.PreimageOracleData) ([]byte, error) {
	return r.fdgAbi.Pack(
		"addLocalData",
		data.OracleKey,
		big.NewInt(0),
	)
}

// buildGlobalOracleData takes the global preimage key and data
// and creates tx data to load the key, data pair into the
// PreimageOracle contract.
//
// Encoded call to: loadKeccak256PreimagePart(uint256 _partOffset, bytes calldata _preimage) external
func (r *faultResponder) buildGlobalOracleData(data types.PreimageOracleData) ([]byte, error) {
	return r.preimageOracleAbi.Pack(
		"loadKeccak256PreimagePart",
		big.NewInt(0),
		data.OracleData,
	)
}

// CanResolve determines if the resolve function on the fault dispute game contract
// would succeed. Returns true if the game can be resolved, otherwise false.
func (r *faultResponder) CanResolve(ctx context.Context) bool {
	txData, err := r.buildResolveData()
	if err != nil {
		return false
	}
	_, err = r.txMgr.Call(ctx, ethereum.CallMsg{
		To:   &r.fdgAddr,
		Data: txData,
	}, nil)
	return err == nil
}

// Resolve executes a resolve transaction to resolve a fault dispute game.
func (r *faultResponder) Resolve(ctx context.Context) error {
	txData, err := r.buildResolveData()
	if err != nil {
		return err
	}

	return r.sendTxAndWait(ctx, txData)
}

// Respond takes a [Claim] and executes the response action.
func (r *faultResponder) Respond(ctx context.Context, response types.Claim) error {
	txData, err := r.BuildTx(ctx, response)
	if err != nil {
		return err
	}
	return r.sendTxAndWait(ctx, txData)
}

// sendTxAndWait sends a transaction through the [txmgr] and waits for a receipt.
// This sets the tx GasLimit to 0, performing gas estimation online through the [txmgr].
func (r *faultResponder) sendTxAndWait(ctx context.Context, txData []byte) error {
	receipt, err := r.txMgr.Send(ctx, txmgr.TxCandidate{
		To:       &r.fdgAddr,
		TxData:   txData,
		GasLimit: 0,
	})
	if err != nil {
		return err
	}
	if receipt.Status == ethtypes.ReceiptStatusFailed {
		r.log.Error("Responder tx successfully published but reverted", "tx_hash", receipt.TxHash)
	} else {
		r.log.Debug("Responder tx successfully published", "tx_hash", receipt.TxHash)
	}
	return nil
}

// buildStepTxData creates the transaction data for the step function.
func (r *faultResponder) buildStepTxData(stepData types.StepCallData) ([]byte, error) {
	return r.fdgAbi.Pack(
		"step",
		big.NewInt(int64(stepData.ClaimIndex)),
		stepData.IsAttack,
		stepData.StateData,
		stepData.Proof,
	)
}

// Step accepts step data and executes the step on the fault dispute game contract.
func (r *faultResponder) Step(ctx context.Context, stepData types.StepCallData) error {
	txData, err := r.buildStepTxData(stepData)
	if err != nil {
		return err
	}
	return r.sendTxAndWait(ctx, txData)
}
