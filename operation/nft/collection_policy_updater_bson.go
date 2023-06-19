package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact CollectionPolicyUpdaterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      fact.Hint().String(),
			"hash":       fact.BaseFact.Hash().String(),
			"token":      fact.BaseFact.Token(),
			"sender":     fact.sender,
			"contract":   fact.contract,
			"collection": fact.collection,
			"name":       fact.name,
			"royalty":    fact.royalty,
			"uri":        fact.uri,
			"whitelist":  fact.whitelist,
			"currency":   fact.currency,
		})
}

type CollectionPolicyUpdaterFactBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Sender     string   `bson:"sender"`
	Contract   string   `bson:"contract"`
	Collection string   `bson:"collection"`
	Name       string   `bson:"name"`
	Royalty    uint     `bson:"royalty"`
	URI        string   `bson:"uri"`
	Whitelist  []string `bson:"whitelist"`
	Currency   string   `bson:"currency"`
}

func (fact *CollectionPolicyUpdaterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of CollectionPolicyUpdaterFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf CollectionPolicyUpdaterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unmarshal(enc, uf.Sender, uf.Contract, uf.Collection, uf.Name, uf.Royalty, uf.URI, uf.Whitelist, uf.Currency)
}

func (op CollectionPolicyUpdater) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *CollectionPolicyUpdater) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of CollectionPolicyUpdater")

	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
