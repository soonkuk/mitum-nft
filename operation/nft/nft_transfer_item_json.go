package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type NFTTransferItemJSONMarshaler struct {
	hint.BaseHinter
	Contract   mitumbase.Address `json:"contract"`
	Collection types.ContractID  `json:"collection"`
	Receiver   mitumbase.Address `json:"receiver"`
	NFTidx     uint64            `json:"nft"`
	Currency   types.CurrencyID  `json:"currency"`
}

func (it NFTTransferItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(NFTTransferItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Contract:   it.contract,
		Collection: it.collection,
		Receiver:   it.receiver,
		NFTidx:     it.nft,
		Currency:   it.currency,
	})
}

type NFTTransferItemJSONUnmarshaler struct {
	Hint       hint.Hint `json:"_hint"`
	Contract   string    `json:"contract"`
	Collection string    `json:"collection"`
	Receiver   string    `json:"receiver"`
	NFTidx     uint64    `json:"nft"`
	Currency   string    `json:"currency"`
}

func (it *NFTTransferItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of NFTTransferItem")

	var u NFTTransferItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return it.unmarshal(enc, u.Hint, u.Contract, u.Collection, u.Receiver, u.NFTidx, u.Currency)
}
