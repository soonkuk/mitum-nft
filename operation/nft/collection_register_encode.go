package nft

import (
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-nft/v2/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CollectionRegisterFact) unmarshal(
	enc encoder.Encoder,
	sd string,
	ca string,
	sb string,
	nm string,
	ry uint,
	uri string,
	bws []string,
	cid string,
) error {
	e := util.StringError("failed to unmarshal CollectionRegisterFact")

	fact.currency = currencytypes.CurrencyID(cid)

	sender, err := mitumbase.DecodeAddress(sd, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.sender = sender

	fact.collection = currencytypes.ContractID(sb)
	fact.name = types.CollectionName(nm)
	fact.royalty = types.PaymentParameter(ry)
	fact.uri = types.URI(uri)

	contract, err := mitumbase.DecodeAddress(ca, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.contract = contract

	whites := make([]mitumbase.Address, len(bws))
	for i, bw := range bws {
		white, err := mitumbase.DecodeAddress(bw, enc)
		if err != nil {
			return e.Wrap(err)
		}
		whites[i] = white

	}
	fact.whitelist = whites

	return nil
}
