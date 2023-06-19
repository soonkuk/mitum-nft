package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type SignerJSONMarshaler struct {
	hint.BaseHinter
	Account base.Address `json:"account"`
	Share   uint         `json:"share"`
	Signed  bool         `json:"signed"`
}

func (sgn Signer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SignerJSONMarshaler{
		BaseHinter: sgn.BaseHinter,
		Account:    sgn.account,
		Share:      sgn.share,
		Signed:     sgn.signed,
	})
}

type SignerJSONUnmarshaler struct {
	Hint    hint.Hint `json:"_hint"`
	Account string    `json:"account"`
	Share   uint      `json:"share"`
	Signed  bool      `json:"signed"`
}

func (sgn *Signer) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of Signer")

	var u SignerJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return sgn.unmarshal(enc, u.Hint, u.Account, u.Share, u.Signed)
}
