package nft

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it NFTSignItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      it.Hint().String(),
			"contract":   it.contract,
			"collection": it.collection,
			"nft":        it.nft,
			"currency":   it.currency,
		},
	)
}

type NFTSignItemBSONUnmarshaler struct {
	Hint       string `bson:"_hint"`
	Contract   string `bson:"contract"`
	Collection string `bson:"collection"`
	NFT        uint64 `bson:"nft"`
	Currency   string `bson:"currency"`
}

func (it *NFTSignItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFTSignItem")

	var u NFTSignItemBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unmarshal(enc, ht, u.Contract, u.Collection, u.NFT, u.Currency)
}
