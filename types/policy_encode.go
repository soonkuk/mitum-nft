package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (p *CollectionPolicy) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	nm string,
	ry uint,
	uri string,
	bws []string,
) error {
	e := util.StringError("failed to unmarshal CollectionPoicy")

	p.BaseHinter = hint.NewBaseHinter(ht)
	p.name = CollectionName(nm)
	p.royalty = PaymentParameter(ry)
	p.uri = URI(uri)

	whites := make([]base.Address, len(bws))
	for i, bw := range bws {
		white, err := base.DecodeAddress(bw, enc)
		if err != nil {
			return e.Wrap(err)
		}
		whites[i] = white
	}
	p.whites = whites

	return nil
}
