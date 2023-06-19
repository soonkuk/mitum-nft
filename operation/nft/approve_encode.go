package nft

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *ApproveFact) unmarshal(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringError("failed to unmarshal ApproveFact")

	sender, err := base.DecodeAddress(sd, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.sender = sender

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]ApproveItem, len(hit))
	for i, hinter := range hit {
		item, ok := hinter.(ApproveItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected ApproveItem, not %T", hinter))
		}

		items[i] = item
	}
	fact.items = items

	return nil
}
