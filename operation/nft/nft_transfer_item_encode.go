package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *NFTTransferItem) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ca, col,
	rc string,
	nid uint64,
	cid string,
) error {
	e := util.StringError("failed to unmarshal NFTTransferItem")

	it.BaseHinter = hint.NewBaseHinter(ht)
	it.collection = types.ContractID(col)
	switch a, err := mitumbase.DecodeAddress(ca, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.contract = a
	}

	receiver, err := mitumbase.DecodeAddress(rc, enc)
	if err != nil {
		return e.Wrap(err)
	}
	it.receiver = receiver
	it.nft = nid
	it.currency = types.CurrencyID(cid)

	return nil
}
