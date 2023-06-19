package nft

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *MintFact) unmarshal(
	enc encoder.Encoder,
	sd string,
	bits []byte,
) error {
	e := util.StringError("failed to unmarshal MintFact")

	switch sender, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = sender
	}

	hits, err := enc.DecodeSlice(bits)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]MintItem, len(hits))
	for i, hinter := range hits {
		item, ok := hinter.(MintItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected MintItem, not %T", hinter))
		}

		items[i] = item
	}
	fact.items = items

	return nil
}
