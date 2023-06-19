package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (sgn *Signer) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ac string,
	sh uint,
	sg bool,
) error {
	e := util.StringError("failed to unmarshal Signer")

	sgn.BaseHinter = hint.NewBaseHinter(ht)
	sgn.share = sh
	sgn.signed = sg

	account, err := base.DecodeAddress(ac, enc)
	if err != nil {
		return e.Wrap(err)
	}
	sgn.account = account

	return nil
}
