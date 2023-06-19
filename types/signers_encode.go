package types

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (sgns *Signers) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	tt uint,
	bsns []byte,
) error {
	e := util.StringError("failed to unmarshal Signers")

	sgns.BaseHinter = hint.NewBaseHinter(ht)
	sgns.total = tt

	hinters, err := enc.DecodeSlice(bsns)
	if err != nil {
		return e.Wrap(err)
	}

	signers := make([]Signer, len(hinters))
	for i, hinter := range hinters {
		signer, ok := hinter.(Signer)
		if !ok {
			return e.Wrap(errors.Errorf("expected Signer, not %T", hinter))
		}

		signers[i] = signer
	}
	sgns.signers = signers

	return nil
}
