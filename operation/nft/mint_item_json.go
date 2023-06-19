package nft

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-nft/v2/types"

	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type MintItemJSONMarshaler struct {
	hint.BaseHinter
	Contract   mitumbase.Address        `json:"contract"`
	Collection currencytypes.ContractID `json:"collection"`
	Hash       types.NFTHash            `json:"hash"`
	Uri        types.URI                `json:"uri"`
	Creators   types.Signers            `json:"creators"`
	Currency   currencytypes.CurrencyID `json:"currency"`
}

func (it MintItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(MintItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Contract:   it.contract,
		Collection: it.collection,
		Hash:       it.hash,
		Uri:        it.uri,
		Creators:   it.creators,
		Currency:   it.currency,
	})
}

type MintItemJSONUnmarshaler struct {
	Hint       hint.Hint       `json:"_hint"`
	Contract   string          `json:"contract"`
	Collection string          `json:"collection"`
	Hash       string          `json:"hash"`
	Uri        string          `json:"uri"`
	Creators   json.RawMessage `json:"creators"`
	Currency   string          `json:"currency"`
}

func (it *MintItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of MintItem")

	var u MintItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return it.unmarshal(enc, u.Hint, u.Contract, u.Collection, u.Hash, u.Uri, u.Creators, u.Currency)
}
