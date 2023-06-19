package types

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type SignersJSONMarshaler struct {
	hint.BaseHinter
	Total   uint     `json:"total"`
	Signers []Signer `json:"signers"`
}

func (sgns Signers) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SignersJSONMarshaler{
		BaseHinter: sgns.BaseHinter,
		Total:      sgns.total,
		Signers:    sgns.signers,
	})
}

type SignersJSONUnmarshaler struct {
	Hint    hint.Hint       `json:"_hint"`
	Total   uint            `json:"total"`
	Signers json.RawMessage `json:"signers"`
}

func (sgns *Signers) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of Signers")

	var u SignersJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return sgns.unmarshal(enc, u.Hint, u.Total, u.Signers)
}
