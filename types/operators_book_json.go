package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type OperatorsBookJSONMarshaler struct {
	hint.BaseHinter
	Collection types.ContractID `json:"collection"`
	Operators  []base.Address   `json:"operators"`
}

func (ob OperatorsBook) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperatorsBookJSONMarshaler{
		BaseHinter: ob.BaseHinter,
		Collection: ob.collection,
		Operators:  ob.operators,
	})
}

type OperatorsBookJSONUnmarshaler struct {
	Hint       hint.Hint `json:"_hint"`
	Collection string    `json:"collection"`
	Operators  []string  `json:"operators"`
}

func (ob *OperatorsBook) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of operators book")

	var u OperatorsBookJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return ob.unmarshal(enc, u.Hint, u.Collection, u.Operators)
}
