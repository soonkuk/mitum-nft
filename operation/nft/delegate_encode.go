package nft

import (
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *DelegateFact) unmarshal(
	enc encoder.Encoder,
	sd string,
	bits []byte,
) error {
	e := util.StringError("failed to unmarshal DelegateFact")

	sender, err := mitumbase.DecodeAddress(sd, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.sender = sender

	hits, err := enc.DecodeSlice(bits)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]DelegateItem, len(hits))
	for i, hinter := range hits {
		item, ok := hinter.(DelegateItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected DelegateItem, not %T", hinter))
		}

		items[i] = item
	}
	fact.items = items

	return nil
}
