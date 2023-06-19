package nft

import (
	"context"
	"sync"

	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	statenft "github.com/ProtoconNet/mitum-nft/v2/state"
	"github.com/ProtoconNet/mitum-nft/v2/types"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var nftSignItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(NFTSignItemProcessor)
	},
}

var nftSignProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(NFTSignProcessor)
	},
}

func (NFTSign) Process(
	ctx context.Context, getStateFunc mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type NFTSignItemProcessor struct {
	h      util.Hash
	sender mitumbase.Address
	item   NFTSignItem
}

func (ipp *NFTSignItemProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) error {
	nid := ipp.item.NFT()

	st, err := state.ExistsState(statenft.NFTStateKey(ipp.item.contract, ipp.item.collection, statenft.CollectionKey), "key of design", getStateFunc)
	if err != nil {
		return errors.Errorf("collection design not found, %q: %w", ipp.item.collection, err)
	}

	design, err := statenft.StateCollectionValue(st)
	if err != nil {
		return errors.Errorf("collection design value not found, %q: %w", ipp.item.collection, err)
	}

	if !design.Active() {
		return errors.Errorf("deactivated collection, %q", ipp.item.collection)
	}
	st, err = state.ExistsState(stateextension.StateKeyContractAccount(ipp.item.contract), "contract account", getStateFunc)
	if err != nil {
		return errors.Errorf("parent not found, %q: %w", design.Parent(), err)
	}

	ca, err := stateextension.StateContractAccountValue(st)
	if err != nil {
		return errors.Errorf("contract account value not found, %q: %w", ipp.item.contract, err)
	}

	if !ca.IsActive() {
		return errors.Errorf("deactivated contract account, %q", ipp.item.contract)
	}

	st, err = state.ExistsState(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, nid), "key of nft", getStateFunc)
	if err != nil {
		return errors.Errorf("nft not found, %q: %w", nid, err)
	}

	nv, err := statenft.StateNFTValue(st)
	if err != nil {
		return errors.Errorf("nft value not found, %q: %w", nid, err)
	}

	if !nv.Active() {
		return errors.Errorf("burned nft, %q", nid)
	}

	if nv.Creators().IsSignedByAddress(ipp.sender) {
		return errors.Errorf("already signed nft, %q-%q", ipp.sender, nv.ID())
	}

	return nil
}

func (ipp *NFTSignItemProcessor) Process(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, error) {
	nid := ipp.item.NFT()

	st, err := state.ExistsState(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, nid), "key of nft", getStateFunc)
	if err != nil {
		return nil, errors.Errorf("nft not found, %q: %w", nid, err)
	}

	nv, err := statenft.StateNFTValue(st)
	if err != nil {
		return nil, errors.Errorf("nft value not found, %q: %w", nid, err)
	}

	signers := nv.Creators()

	idx := signers.IndexByAddress(ipp.sender)
	if idx < 0 {
		return nil, errors.Errorf("not signer of nft, %q-%q", ipp.sender, nv.ID())
	}

	signer := types.NewSigner(signers.Signers()[idx].Account(), signers.Signers()[idx].Share(), true)
	if err := signer.IsValid(nil); err != nil {
		return nil, errors.Errorf("invalid signer, %q", signer.Account())
	}

	sns := &signers
	if err := sns.SetSigner(signer); err != nil {
		return nil, errors.Errorf("failed to set signer for signers, %q: %w", signer, err)
	}

	n := types.NewNFT(nv.ID(), nv.Active(), nv.Owner(), nv.NFTHash(), nv.URI(), nv.Approved(), *sns)

	if err := n.IsValid(nil); err != nil {
		return nil, errors.Errorf("invalid nft, %q: %w", n.ID(), err)
	}

	sts := make([]mitumbase.StateMergeValue, 1)

	sts[0] = state.NewStateMergeValue(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, n.ID()), statenft.NewNFTStateValue(n))

	return sts, nil
}

func (ipp *NFTSignItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = NFTSignItem{}
	nftSignItemProcessorPool.Put(ipp)

	return nil
}

type NFTSignProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewNFTSignProcessor() currencytypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new NFTSignProcessor")

		nopp := nftSignProcessorPool.Get()
		opp, ok := nopp.(*NFTSignProcessor)
		if !ok {
			return nil, e.Errorf("expected NFTSignProcessor, not %T", nopp)
		}

		b, err := mitumbase.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *NFTSignProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess NFTSign")

	fact, ok := op.Fact().(NFTSignFact)
	if !ok {
		return ctx, nil, e.Errorf("expected NFTSignFact, not %T", op.Fact())
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := state.CheckNotExistsState(stateextension.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot sign nfts, %q", fact.Sender()), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	for _, item := range fact.Items() {
		ip := nftSignItemProcessorPool.Get()
		ipc, ok := ip.(*NFTSignItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected NFTSignItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess NFTSignItem: %w", err), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *NFTSignProcessor) Process( // nolint:dupl
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process NFTSign")

	fact, ok := op.Fact().(NFTSignFact)
	if !ok {
		return nil, nil, e.Errorf("expected NFTSignFact, not %T", op.Fact())
	}

	var sts []mitumbase.StateMergeValue

	for _, item := range fact.Items() {
		ip := nftSignItemProcessorPool.Get()
		ipc, ok := ip.(*NFTSignItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected NFTSignItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		s, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process MintItem: %w", err), nil
		}
		sts = append(sts, s...)

		ipc.Close()
	}

	fitems := fact.Items()
	items := make([]CollectionItem, len(fitems))
	for i := range fact.Items() {
		items[i] = fitems[i]
	}

	required, err := CalculateCollectionItemsFee(getStateFunc, items)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	}
	sb, err := currency.CheckEnoughBalance(fact.sender, required, getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check enough balance: %w", err), nil
	}

	for i := range sb {
		v, ok := sb[i].Value().(statecurrency.BalanceStateValue)
		if !ok {
			return nil, nil, e.Errorf("expected BalanceStateValue, not %T", sb[i].Value())
		}
		stv := statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(required[i][0])))
		sts = append(sts, state.NewStateMergeValue(sb[i].Key(), stv))
	}

	return sts, nil, nil
}

func (opp *NFTSignProcessor) Close() error {
	nftSignProcessorPool.Put(opp)

	return nil
}
