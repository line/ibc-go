package mock

import (
	cryptocodec "github.com/Finschia/finschia-sdk/crypto/codec"
	"github.com/Finschia/finschia-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/Finschia/finschia-sdk/crypto/types"
	"github.com/Finschia/ostracon/crypto"
	tmtypes "github.com/Finschia/ostracon/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var _ tmtypes.PrivValidator = PV{}

// MockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type PV struct {
	PrivKey cryptotypes.PrivKey
}

func NewPV() PV {
	return PV{ed25519.GenPrivKey()}
}

// GetPubKey implements PrivValidator interface
func (pv PV) GetPubKey() (crypto.PubKey, error) {
	return cryptocodec.ToOcPubKeyInterface(pv.PrivKey.PubKey())
}

// SignVote implements PrivValidator interface
func (pv PV) SignVote(chainID string, vote *tmproto.Vote) error {
	signBytes := tmtypes.VoteSignBytes(chainID, vote)
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig
	return nil
}

// SignProposal implements PrivValidator interface
func (pv PV) SignProposal(chainID string, proposal *tmproto.Proposal) error {
	signBytes := tmtypes.ProposalSignBytes(chainID, proposal)
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = sig
	return nil
}

func (pv PV) GenerateVRFProof(message []byte) (crypto.Proof, error) {
	return nil, nil
}
