package nft

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-nft/v2/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	CollectionRegisterFactHint = hint.MustNewHint("mitum-nft-collection-register-operation-fact-v0.0.1")
	CollectionRegisterHint     = hint.MustNewHint("mitum-nft-collection-register-operation-v0.0.1")
)

type CollectionRegisterFact struct {
	mitumbase.BaseFact
	sender     mitumbase.Address
	contract   mitumbase.Address
	collection currencytypes.ContractID
	name       types.CollectionName
	royalty    types.PaymentParameter
	uri        types.URI
	whitelist  []mitumbase.Address
	currency   currencytypes.CurrencyID
}

func NewCollectionRegisterFact(
	token []byte,
	sender mitumbase.Address,
	contract mitumbase.Address,
	collection currencytypes.ContractID,
	name types.CollectionName,
	royalty types.PaymentParameter,
	uri types.URI,
	whitelist []mitumbase.Address,
	currency currencytypes.CurrencyID,
) CollectionRegisterFact {
	bf := mitumbase.NewBaseFact(CollectionRegisterFactHint, token)
	fact := CollectionRegisterFact{
		BaseFact:   bf,
		sender:     sender,
		contract:   contract,
		collection: collection,
		name:       name,
		royalty:    royalty,
		uri:        uri,
		whitelist:  whitelist,
		currency:   currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CollectionRegisterFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.collection,
		fact.name,
		fact.royalty,
		fact.uri,
		fact.currency,
	); err != nil {
		return err
	}

	if fact.sender.Equal(fact.contract) {
		return util.ErrInvalid.Errorf("sender and contract are the same, %q == %q", fact.sender, fact.contract)
	}

	return nil
}

func (fact CollectionRegisterFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CollectionRegisterFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CollectionRegisterFact) Bytes() []byte {
	as := make([][]byte, len(fact.whitelist))
	for i, white := range fact.whitelist {
		as[i] = white.Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		fact.collection.Bytes(),
		fact.name.Bytes(),
		fact.royalty.Bytes(),
		fact.uri.Bytes(),
		fact.currency.Bytes(),
		util.ConcatBytesSlice(as...),
	)
}

func (fact CollectionRegisterFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact CollectionRegisterFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact CollectionRegisterFact) Contract() mitumbase.Address {
	return fact.contract
}

func (fact CollectionRegisterFact) Collection() currencytypes.ContractID {
	return fact.collection
}

func (fact CollectionRegisterFact) Name() types.CollectionName {
	return fact.name
}

func (fact CollectionRegisterFact) Royalty() types.PaymentParameter {
	return fact.royalty
}

func (fact CollectionRegisterFact) URI() types.URI {
	return fact.uri
}

func (fact CollectionRegisterFact) Whites() []mitumbase.Address {
	return fact.whitelist
}

func (fact CollectionRegisterFact) Addresses() ([]mitumbase.Address, error) {
	l := 2 + len(fact.whitelist)

	as := make([]mitumbase.Address, l)
	copy(as, fact.whitelist)

	as[l-2] = fact.sender
	as[l-1] = fact.contract

	return as, nil
}

func (fact CollectionRegisterFact) Currency() currencytypes.CurrencyID {
	return fact.currency
}

type CollectionRegister struct {
	common.BaseOperation
}

func NewCollectionRegister(fact CollectionRegisterFact) (CollectionRegister, error) {
	return CollectionRegister{BaseOperation: common.NewBaseOperation(CollectionRegisterHint, fact)}, nil
}

func (op *CollectionRegister) HashSign(priv mitumbase.Privatekey, networkID mitumbase.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}