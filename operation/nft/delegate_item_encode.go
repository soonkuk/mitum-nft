package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *DelegateItem) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ca, col string,
	ag string,
	md string,
	cid string,
) error {
	e := util.StringError("failed to unmarshal DelegateItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	it.collection = types.ContractID(col)
	it.mode = DelegateMode(md)
	it.currency = types.CurrencyID(cid)

	switch a, err := mitumbase.DecodeAddress(ca, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.contract = a
	}

	operator, err := mitumbase.DecodeAddress(ag, enc)
	if err != nil {
		return e.Wrap(err)
	}
	it.operator = operator

	return nil
}
