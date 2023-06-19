package types

import (
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type NFTBoxJSONMarshaler struct {
	hint.BaseHinter
	NFTs []uint64 `json:"nfts"`
}

func (nbx NFTBox) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(NFTBoxJSONMarshaler{
		BaseHinter: nbx.BaseHinter,
		NFTs:       nbx.nfts,
	})
}

type NFTBoxJSONUnmarshaler struct {
	Hint hint.Hint `json:"_hint"`
	NFTs []uint64  `json:"nfts"`
}

func (nbx *NFTBox) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of NFTBox")

	var u NFTBoxJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {

		return e.Wrap(err)
	}

	return nbx.unmarshal(enc, u.Hint, u.NFTs)
}
