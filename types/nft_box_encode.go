package types

import (
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (nbx *NFTBox) unmarshal(
	enc encoder.Encoder,
	ht hint.Hint,
	ns []uint64,
) error {
	//e := util.StringError("failed to unmarshal NFTBox")

	nbx.BaseHinter = hint.NewBaseHinter(ht)

	//hns, err := enc.DecodeSlice(bns)
	//if err != nil {
	//	return e.Wrap(err)
	//}

	//nfts := make([]nft.NFTID, len(hns))
	//for i, hinter := range hns {
	//	n, ok := hinter.(nft.NFTID)
	//	if !ok {
	//		return e(errors.Errorf("expected NFTID, not %T", hinter), "")
	//	}
	//
	//	nfts[i] = n
	//
	//var nfts []nft.NFTID
	//for _, n := range ns {
	//	id, err := strconv.ParseUint(n, 10, 64)
	//	if err != nil {
	//		return err
	//	}
	//	nfts = append(nfts, nft.NFTID(id))
	//}

	nbx.nfts = ns

	return nil
}
