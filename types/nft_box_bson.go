package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (nbx NFTBox) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": nbx.Hint().String(),
			"nfts":  nbx.nfts,
		},
	)
}

type NFTBoxBSONUnmarshaler struct {
	Hint string   `bson:"_hint"`
	NFTs []uint64 `bson:"nfts"`
}

func (nbx *NFTBox) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFTBox")

	var u NFTBoxBSONUnmarshaler
	if err := bsonenc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return nbx.unmarshal(enc, ht, u.NFTs)
}
