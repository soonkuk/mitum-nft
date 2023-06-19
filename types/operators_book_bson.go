package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (ob OperatorsBook) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":      ob.Hint().String(),
		"collection": ob.collection,
		"operators":  ob.operators,
	})
}

type OperatorsBookBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Collection string   `bson:"collection"`
	Operators  []string `bson:"operators"`
}

func (ob *OperatorsBook) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of operators book")

	var u OperatorsBookBSONUnmarshaler
	if err := bsonenc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return ob.unmarshal(enc, ht, u.Collection, u.Operators)
}
