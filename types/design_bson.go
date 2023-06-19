package types

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (de Design) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      de.Hint().String(),
			"parent":     de.parent,
			"creator":    de.creator,
			"collection": de.collection,
			"active":     de.active,
			"policy":     de.policy,
		})
}

type DesignBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Parent     string   `bson:"parent"`
	Creator    string   `bson:"creator"`
	Collection string   `bson:"collection"`
	Active     bool     `bson:"active"`
	Policy     bson.Raw `bson:"policy"`
}

func (de *Design) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Design")

	var u DesignBSONUnmarshaler
	if err := bson.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return de.unmarshal(enc, ht, u.Parent, u.Creator, u.Collection, u.Active, u.Policy)
}
