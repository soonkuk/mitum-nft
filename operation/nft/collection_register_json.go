package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-nft/v2/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type CollectionRegisterFactJSONMarshaler struct {
	mitumbase.BaseFactJSONMarshaler
	Sender     mitumbase.Address        `json:"sender"`
	Contract   mitumbase.Address        `json:"contract"`
	Collection currencytypes.ContractID `json:"collection"`
	Name       types.CollectionName     `json:"name"`
	Royalty    types.PaymentParameter   `json:"royalty"`
	URI        types.URI                `json:"uri"`
	Whites     []mitumbase.Address      `json:"whites"`
	Currency   currencytypes.CurrencyID `json:"currency"`
}

func (fact CollectionRegisterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CollectionRegisterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		Collection:            fact.collection,
		Name:                  fact.name,
		Royalty:               fact.royalty,
		URI:                   fact.uri,
		Whites:                fact.whitelist,
		Currency:              fact.currency,
	})
}

type CollectionRegisterFactJSONUnmarshaler struct {
	mitumbase.BaseFactJSONUnmarshaler
	Sender     string   `json:"sender"`
	Contract   string   `json:"contract"`
	Collection string   `json:"collection"`
	Name       string   `json:"name"`
	Royalty    uint     `json:"royalty"`
	URI        string   `json:"uri"`
	Whites     []string `json:"whites"`
	Currency   string   `json:"currency"`
}

func (fact *CollectionRegisterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CollectionRegisterFact")

	var u CollectionRegisterFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	return fact.unmarshal(enc, u.Sender, u.Contract, u.Collection, u.Name, u.Royalty, u.URI, u.Whites, u.Currency)
}

type collectionRegisterMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op CollectionRegister) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(collectionRegisterMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CollectionRegister) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CurrecyRegister")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
