package nft

import (
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-nft/v2/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CollectionPolicyUpdaterFact) unmarshal(
	enc encoder.Encoder,
	sd string,
	ct string,
	col string,
	nm string,
	ry uint,
	uri string,
	bws []string,
	cid string,
) error {
	e := util.StringError("failed to unmarshal CollectionPolicyUpdaterFact")

	fact.collection = currencytypes.ContractID(col)
	fact.currency = currencytypes.CurrencyID(cid)

	sender, err := mitumbase.DecodeAddress(sd, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.sender = sender

	contract, err := mitumbase.DecodeAddress(sd, enc)
	if err != nil {
		return e.Wrap(err)
	}
	fact.contract = contract

	fact.name = types.CollectionName(nm)
	fact.royalty = types.PaymentParameter(ry)
	fact.uri = types.URI(uri)

	//hits, err := enc.DecodeSlice(bws)
	//if err != nil {
	//	return e.Wrap(err)
	//}
	//
	//whitelist := make([]mitumbase.Address, len(bws))
	//for i := range hits {
	//	ad := fmt.Sprintf("%v", hits[i])
	//	white, err := mitumbase.DecodeAddress(ad, enc)
	//	if err != nil {
	//		return e.Wrap(err)
	//	}
	//	whitelist[i] = white
	//}
	//fact.whitelist = whitelist
	whitelist := make([]mitumbase.Address, len(bws))
	for i, bw := range bws {
		white, err := mitumbase.DecodeAddress(bw, enc)
		if err != nil {
			return e.Wrap(err)
		}
		whitelist[i] = white
	}
	fact.whitelist = whitelist

	return nil
}
