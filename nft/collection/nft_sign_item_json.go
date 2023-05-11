package collection

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/v2/currency"
	"github.com/ProtoconNet/mitum-nft/nft"

	"github.com/ProtoconNet/mitum-currency/v2/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type NFTSignItemJSONMarshaler struct {
	hint.BaseHinter
	Contract      base.Address                 `json:"contract"`
	Collection    extensioncurrency.ContractID `json:"collection"`
	Qualification Qualification                `json:"qualification"`
	NFT           nft.NFTID                    `json:"nft"`
	Currency      currency.CurrencyID          `json:"currency"`
}

func (it NFTSignItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(NFTSignItemJSONMarshaler{
		BaseHinter:    it.BaseHinter,
		Contract:      it.contract,
		Collection:    it.collection,
		Qualification: it.qualification,
		NFT:           it.nft,
		Currency:      it.currency,
	})
}

type NFTSignItemJSONUnmarshaler struct {
	Hint          hint.Hint       `json:"_hint"`
	Contract      string          `json:"contract"`
	Collection    string          `json:"collection"`
	Qualification string          `json:"qualification"`
	NFT           json.RawMessage `json:"nft"`
	Currency      string          `json:"currency"`
}

func (it *NFTSignItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of NFTSignItem")

	var u NFTSignItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	return it.unmarshal(enc, u.Hint, u.Contract, u.Collection, u.Qualification, u.NFT, u.Currency)
}
