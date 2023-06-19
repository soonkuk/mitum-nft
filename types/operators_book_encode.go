package types

import (
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (ob *OperatorsBook) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	col string,
	bags []string,
) error {
	e := util.StringError("failed to unmarshal operators book")

	ob.BaseHinter = hint.NewBaseHinter(ht)
	ob.collection = currencytypes.ContractID(col)

	operators := make([]mitumbase.Address, len(bags))
	for i, bag := range bags {
		operator, err := mitumbase.DecodeAddress(bag, enc)
		if err != nil {
			return e.Wrap(err)
		}
		operators[i] = operator
	}
	ob.operators = operators

	return nil
}
