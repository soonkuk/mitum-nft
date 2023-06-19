package nft

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it NFTTransferItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      it.Hint().String(),
			"contract":   it.contract,
			"collection": it.collection,
			"receiver":   it.receiver,
			"nft":        it.nft,
			"currency":   it.currency,
		},
	)
}

type NFTTransferItemBSONUnmarshaler struct {
	Hint       string `bson:"_hint"`
	Contract   string `bson:"contract"`
	Collection string `bson:"collection"`
	Receiver   string `bson:"receiver"`
	NFTidx     uint64 `bson:"nft"`
	Currency   string `bson:"currency"`
}

func (it *NFTTransferItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFTTransferItem")

	var u NFTTransferItemBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unmarshal(enc, ht, u.Contract, u.Collection, u.Receiver, u.NFTidx, u.Currency)
}
