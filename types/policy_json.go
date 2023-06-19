package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CollectionPolicyJSONMarshaler struct {
	hint.BaseHinter
	Name    CollectionName   `json:"name"`
	Royalty PaymentParameter `json:"royalty"`
	URI     URI              `json:"uri"`
	Whites  []base.Address   `json:"whites"`
}

func (p CollectionPolicy) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CollectionPolicyJSONMarshaler{
		BaseHinter: p.BaseHinter,
		Name:       p.name,
		Royalty:    p.royalty,
		URI:        p.uri,
		Whites:     p.whites,
	})
}

type CollectionPolicyJSONUnmarshaler struct {
	Hint    hint.Hint `json:"_hint"`
	Name    string    `json:"name"`
	Royalty uint      `json:"royalty"`
	URI     string    `json:"uri"`
	Whites  []string  `json:"whites"`
}

func (p *CollectionPolicy) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CollectionPolicy")

	var u CollectionPolicyJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return p.unmarshal(enc, u.Hint, u.Name, u.Royalty, u.URI, u.Whites)
}
