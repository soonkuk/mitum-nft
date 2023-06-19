package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (sgn Signer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":   sgn.Hint().String(),
		"account": sgn.account,
		"share":   sgn.share,
		"signed":  sgn.signed,
	})
}

type SignerBSONUnmarshaler struct {
	Hint    string `bson:"_hint"`
	Account string `bson:"account"`
	Share   uint   `bson:"share"`
	Signed  bool   `bson:"signed"`
}

func (sgn *Signer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Signer")

	var u SignerBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return sgn.unmarshal(enc, ht, u.Account, u.Share, u.Signed)
}
