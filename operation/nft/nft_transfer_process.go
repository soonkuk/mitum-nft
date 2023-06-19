package nft

import (
	"context"
	"sync"

	statenft "github.com/ProtoconNet/mitum-nft/v2/state"
	"github.com/ProtoconNet/mitum-nft/v2/types"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var nftTransferItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(NFTTransferItemProcessor)
	},
}

var nftTransferProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(NFTTransferProcessor)
	},
}

func (NFTTransfer) Process(
	ctx context.Context, getStateFunc mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type NFTTransferItemProcessor struct {
	h      util.Hash
	sender mitumbase.Address
	item   NFTTransferItem
}

func (ipp *NFTTransferItemProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) error {
	receiver := ipp.item.Receiver()

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(receiver), getStateFunc); err != nil {
		return errors.Errorf("receiver not found, %q: %w", receiver, err)
	}

	if err := state.CheckNotExistsState(stateextension.StateKeyContractAccount(receiver), getStateFunc); err != nil {
		return errors.Errorf("contract account cannot receive nfts, %q: %w", receiver, err)
	}

	nid := ipp.item.NFT()

	st, err := state.ExistsState(statenft.NFTStateKey(ipp.item.contract, ipp.item.collection, statenft.CollectionKey), "design", getStateFunc)
	if err != nil {
		return errors.Errorf("collection design not found, %q: %w", ipp.item.collection, err)
	}

	design, err := statenft.StateCollectionValue(st)
	if err != nil {
		return errors.Errorf("collection design not found, %q: %w", ipp.item.collection, err)
	}
	if !design.Active() {
		return errors.Errorf("deactivated collection, %q", design.Collection())
	}

	st, err = state.ExistsState(stateextension.StateKeyContractAccount(design.Parent()), "key of contract account", getStateFunc)
	if err != nil {
		return errors.Errorf("parent not found, %q: %w", design.Parent(), err)
	}

	ca, err := stateextension.StateContractAccountValue(st)
	if err != nil {
		return errors.Errorf("parent account value not found, %q: %w", design.Parent(), err)
	}

	if !ca.IsActive() {
		return errors.Errorf("deactivated contract account, %q", design.Parent())
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

	if !(nv.Owner().Equal(ipp.sender) || nv.Approved().Equal(ipp.sender)) {
		if st, err := state.ExistsState(statenft.StateKeyOperators(ipp.item.contract, ipp.item.collection, nv.Owner()), "operators", getStateFunc); err != nil {
			return errors.Errorf("unauthorized sender, %q: %w", ipp.sender, err)
		} else if box, err := statenft.StateOperatorsBookValue(st); err != nil {
			return errors.Errorf("operator book value not found, %q: %w", ipp.sender, err)
		} else if !box.Exists(ipp.sender) {
			return errors.Errorf("unauthorized sender, %q", ipp.sender)
		}
	}

	return nil
}

func (ipp *NFTTransferItemProcessor) Process(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, error) {
	receiver := ipp.item.Receiver()
	nid := ipp.item.NFT()

	st, err := state.ExistsState(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, nid), "key of nft", getStateFunc)
	if err != nil {
		return nil, errors.Errorf("nft not found, %q: %w", nid, err)
	}

	nv, err := statenft.StateNFTValue(st)
	if err != nil {
		return nil, errors.Errorf("nft value not found, %q: %w", nid, err)
	}

	n := types.NewNFT(nid, nv.Active(), receiver, nv.NFTHash(), nv.URI(), receiver, nv.Creators())
	if err := n.IsValid(nil); err != nil {
		return nil, errors.Errorf("invalid nft, %q: %w", nid, err)
	}

	sts := make([]mitumbase.StateMergeValue, 1)

	sts[0] = state.NewStateMergeValue(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, ipp.item.NFT()), statenft.NewNFTStateValue(n))

	return sts, nil
}

func (ipp *NFTTransferItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = NFTTransferItem{}

	nftTransferItemProcessorPool.Put(ipp)

	return nil
}

type NFTTransferProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewNFTTransferProcessor() currencytypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new NFTTransferProcessor")

		nopp := nftTransferProcessorPool.Get()
		opp, ok := nopp.(*NFTTransferProcessor)
		if !ok {
			return nil, e.Errorf("expected NFTTransferProcessor, not %T", nopp)
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

func (opp *NFTTransferProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess NFTTransfer")

	fact, ok := op.Fact().(NFTTransferFact)
	if !ok {
		return ctx, nil, e.Errorf("expected NFTTransferFact, not %T", op.Fact())
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := state.CheckNotExistsState(stateextension.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot transfer nfts, %q", fact.Sender()), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	for _, item := range fact.Items() {
		ip := nftTransferItemProcessorPool.Get()
		ipc, ok := ip.(*NFTTransferItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected NFTTransferItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess NFTTransferItem: %w", err), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *NFTTransferProcessor) Process( // nolint:dupl
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process NFTTransfer")

	fact, ok := op.Fact().(NFTTransferFact)
	if !ok {
		return nil, nil, e.Errorf("expected NFTTransferFact, not %T", op.Fact())
	}

	var sts []mitumbase.StateMergeValue // nolint:prealloc
	for _, item := range fact.Items() {
		ip := nftTransferItemProcessorPool.Get()
		ipc, ok := ip.(*NFTTransferItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected NFTTransferItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		s, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process NFTTransferItem: %w", err), nil
		}
		sts = append(sts, s...)

		ipc.Close()
	}

	required, err := opp.calculateItemsFee(op, getStateFunc)
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

func (opp *NFTTransferProcessor) Close() error {
	nftTransferProcessorPool.Put(opp)

	return nil
}

func (opp *NFTTransferProcessor) calculateItemsFee(op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (map[currencytypes.CurrencyID][2]common.Big, error) {
	fact, ok := op.Fact().(NFTTransferFact)
	if !ok {
		return nil, errors.Errorf("expected NFTTransferFact, not %T", op.Fact())
	}

	items := make([]CollectionItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return CalculateCollectionItemsFee(getStateFunc, items)
}
