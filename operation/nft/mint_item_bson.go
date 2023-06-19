package nft

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it MintItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      it.Hint().String(),
			"contract":   it.contract,
			"collection": it.collection,
			"hash":       it.hash,
			"uri":        it.uri,
			"creators":   it.creators,
			"currency":   it.currency,
		},
	)
}

type MintItemBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Contract   string   `bson:"contract"`
	Collection string   `bson:"collection"`
	Hash       string   `bson:"hash"`
	Uri        string   `bson:"uri"`
	Creators   bson.Raw `bson:"creators"`
	Currency   string   `bson:"currency"`
}

func (it *MintItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of MintItem")

	var u MintItemBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unmarshal(enc, ht, u.Contract, u.Collection, u.Hash, u.Uri, u.Creators, u.Currency)
}
