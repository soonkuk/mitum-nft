package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *ApproveItem) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ca, col,
	ap string,
	idx uint64,
	cid string,
) error {
	e := util.StringError("failed to unmarshal ApproveItem")

	it.BaseHinter = hint.NewBaseHinter(ht)
	it.currency = types.CurrencyID(cid)
	it.collection = types.ContractID(col)
	switch a, err := mitumbase.DecodeAddress(ca, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.contract = a
	}

	approved, err := mitumbase.DecodeAddress(ap, enc)
	if err != nil {
		return e.Wrap(err)
	}
	it.approved = approved
	it.idx = idx

	return nil
}
