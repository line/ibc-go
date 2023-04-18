package types

import (
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdkerrors "github.com/line/lbm-sdk/types/errors"
	octypes "github.com/line/ostracon/types"

	clienttypes "github.com/Finschia/ibc-go/v3/modules/core/02-client/types"
	host "github.com/Finschia/ibc-go/v3/modules/core/24-host"
	"github.com/Finschia/ibc-go/v3/modules/core/exported"
)

var _ exported.Misbehaviour = &Misbehaviour{}

// FrozenHeight is same for all misbehaviour
var FrozenHeight = clienttypes.NewHeight(0, 1)

// NewMisbehaviour creates a new Misbehaviour instance.
func NewMisbehaviour(clientID string, header1, header2 *Header) *Misbehaviour {
	return &Misbehaviour{
		ClientId: clientID,
		Header1:  header1,
		Header2:  header2,
	}
}

// ClientType is Tendermint light client
func (misbehaviour Misbehaviour) ClientType() string {
	return exported.Tendermint
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (misbehaviour Misbehaviour) GetClientID() string {
	return misbehaviour.ClientId
}

// GetTime returns the timestamp at which misbehaviour occurred. It uses the
// maximum value from both headers to prevent producing an invalid header outside
// of the misbehaviour age range.
func (misbehaviour Misbehaviour) GetTime() time.Time {
	t1, t2 := misbehaviour.Header1.GetTime(), misbehaviour.Header2.GetTime()
	if t1.After(t2) {
		return t1
	}
	return t2
}

// ValidateBasic implements Misbehaviour interface
func (misbehaviour Misbehaviour) ValidateBasic() error {
	if misbehaviour.Header1 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "misbehaviour Header1 cannot be nil")
	}
	if misbehaviour.Header2 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "misbehaviour Header2 cannot be nil")
	}
	if misbehaviour.Header1.TrustedHeight.RevisionHeight == 0 {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "misbehaviour Header1 cannot have zero revision height")
	}
	if misbehaviour.Header2.TrustedHeight.RevisionHeight == 0 {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "misbehaviour Header2 cannot have zero revision height")
	}
	if misbehaviour.Header1.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header1 cannot be empty")
	}
	if misbehaviour.Header2.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header2 cannot be empty")
	}
	if misbehaviour.Header1.Header.ChainID != misbehaviour.Header2.Header.ChainID {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "headers must have identical chainIDs")
	}

	if err := host.ClientIdentifierValidator(misbehaviour.ClientId); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour client ID is invalid")
	}

	// ValidateBasic on both validators
	if err := misbehaviour.Header1.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidMisbehaviour,
			sdkerrors.Wrap(err, "header 1 failed validation").Error(),
		)
	}
	if err := misbehaviour.Header2.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidMisbehaviour,
			sdkerrors.Wrap(err, "header 2 failed validation").Error(),
		)
	}
	// Ensure that Height1 is greater than or equal to Height2
	if misbehaviour.Header1.GetHeight().LT(misbehaviour.Header2.GetHeight()) {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidMisbehaviour, "Header1 height is less than Header2 height (%s < %s)", misbehaviour.Header1.GetHeight(), misbehaviour.Header2.GetHeight())
	}

	blockID1, err := octypes.BlockIDFromProto(&misbehaviour.Header1.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 1 in misbehaviour")
	}
	blockID2, err := octypes.BlockIDFromProto(&misbehaviour.Header2.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 2 in misbehaviour")
	}

	if err := validCommit(misbehaviour.Header1.Header.ChainID, *blockID1,
		misbehaviour.Header1.Commit, misbehaviour.Header1.ValidatorSet); err != nil {
		return err
	}
	if err := validCommit(misbehaviour.Header2.Header.ChainID, *blockID2,
		misbehaviour.Header2.Commit, misbehaviour.Header2.ValidatorSet); err != nil {
		return err
	}
	return nil
}

// validCommit checks if the given commit is a valid commit from the passed-in validatorset
func validCommit(chainID string, blockID octypes.BlockID, commit *tmproto.Commit, valSet *tmproto.ValidatorSet) (err error) {
	tmCommit, err := octypes.CommitFromProto(commit)
	if err != nil {
		return sdkerrors.Wrap(err, "commit is not tendermint commit type")
	}
	tmValset, err := octypes.ValidatorSetFromProto(valSet)
	if err != nil {
		return sdkerrors.Wrap(err, "validator set is not tendermint validator set type")
	}

	if err := tmValset.VerifyCommitLight(chainID, blockID, tmCommit.Height, tmCommit); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "validator set did not commit to header")
	}

	return nil
}
