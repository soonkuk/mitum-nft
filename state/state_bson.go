package state

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-nft/v2/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (s CollectionStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":            s.Hint().String(),
			"collectiondesign": s.Design,
		},
	)
}

type CollectionStateValueBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Design bson.Raw `bson:"collectiondesign"`
}

func (s *CollectionStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of CollectionStateValue")

	var u CollectionStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var nd types.Design
	if err := nd.DecodeBSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	s.Design = nd

	return nil
}

func (s LastNFTIndexStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": s.Hint().String(),
			"index": s.id,
		},
	)
}

type CollectionLastNFTIndexStateValueBSONUnmarshaler struct {
	Hint  string `bson:"_hint"`
	Index uint64 `bson:"index"`
}

func (s *LastNFTIndexStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of CollectionLastNFTIndexStateValue")

	var u CollectionLastNFTIndexStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	s.id = u.Index

	return nil
}

func (s NFTStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": s.Hint().String(),
			"nft":   s.NFT,
		},
	)
}

type NFTStateValueBSONUnmarshaler struct {
	Hint string   `bson:"_hint"`
	NFT  bson.Raw `bson:"nft"`
}

func (s *NFTStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFTStateValue")

	var u NFTStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var n types.NFT
	if err := n.DecodeBSON(u.NFT, enc); err != nil {
		return e.Wrap(err)
	}
	s.NFT = n

	return nil
}

func (s NFTBoxStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  s.Hint().String(),
			"nftbox": s.Box,
		},
	)
}

type NFTBoxStateValueBSONUnmarshaler struct {
	Hint string   `bson:"_hint"`
	Box  bson.Raw `bson:"nftbox"`
}

func (s *NFTBoxStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of NFTBoxStateValue")

	var u NFTBoxStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var box types.NFTBox
	if err := box.DecodeBSON(u.Box, enc); err != nil {
		return e.Wrap(err)
	}
	s.Box = box

	return nil
}

func (s OperatorsBookStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":         s.Hint().String(),
			"operatorsbook": s.Operators,
		},
	)
}

type OperatorsBookStateValueBSONUnmarshaler struct {
	Hint      string   `bson:"_hint"`
	Operators bson.Raw `bson:"operatorsbook"`
}

func (s *OperatorsBookStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of OperatorsBookStateValue")

	var u OperatorsBookStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var operators types.OperatorsBook
	if err := operators.DecodeBSON(u.Operators, enc); err != nil {
		return e.Wrap(err)
	}
	s.Operators = operators

	return nil
}
