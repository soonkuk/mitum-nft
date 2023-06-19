package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (sgns Signers) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   sgns.Hint().String(),
			"total":   sgns.total,
			"signers": sgns.signers,
		})
}

type SignersBSONUnmarshaler struct {
	Hint    string   `bson:"_hint"`
	Total   uint     `bson:"total"`
	Signers bson.Raw `bson:"signers"`
}

func (sgns *Signers) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Signers")

	var u SignersBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return sgns.unmarshal(enc, ht, u.Total, u.Signers)
}
