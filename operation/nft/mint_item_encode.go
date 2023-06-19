package nft

import (
	"github.com/ProtoconNet/mitum-nft/v2/types"
	"github.com/pkg/errors"

	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *MintItem) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ca, col, hs, uri string,
	bcr []byte,
	cid string,
) error {
	e := util.StringError("failed to unmarshal MintItem")

	it.BaseHinter = hint.NewBaseHinter(ht)
	it.collection = currencytypes.ContractID(col)
	it.hash = types.NFTHash(hs)
	it.uri = types.URI(uri)

	switch a, err := mitumbase.DecodeAddress(ca, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.contract = a
	}

	if hinter, err := enc.Decode(bcr); err != nil {
		return e.Wrap(err)
	} else if creators, ok := hinter.(types.Signers); !ok {
		return e.Wrap(errors.Errorf("expected Signers, not %T", hinter))
	} else {
		it.creators = creators
	}

	it.currency = currencytypes.CurrencyID(cid)

	return nil
}
