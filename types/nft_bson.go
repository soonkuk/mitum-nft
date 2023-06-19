package types

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (n NFT) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":    n.Hint().String(),
		"id":       n.id,
		"active":   n.active,
		"owner":    n.owner,
		"hash":     n.hash,
		"uri":      n.uri,
		"approved": n.approved,
		"creators": n.creators,
	})
}

type NFTBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	ID       uint64   `bson:"id"`
	Active   bool     `bson:"active"`
	Owner    string   `bson:"owner"`
	Hash     string   `bson:"hash"`
	URI      string   `bson:"uri"`
	Approved string   `bson:"approved"`
	Creators bson.Raw `bson:"creators"`
}

func (n *NFT) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFT")

	var u NFTBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return n.unmarshal(enc, ht, u.ID, u.Active, u.Owner, u.Hash, u.URI, u.Approved, u.Creators)
}
