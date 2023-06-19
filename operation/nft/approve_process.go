package nft

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-nft/v2/types"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	statenft "github.com/ProtoconNet/mitum-nft/v2/state"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var approveItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(ApproveItemProcessor)
	},
}

var approveProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(ApproveProcessor)
	},
}

func (Approve) Process(
	ctx context.Context, getStateFunc mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type ApproveItemProcessor struct {
	h      util.Hash
	sender mitumbase.Address
	item   ApproveItem
}

func (ipp *ApproveItemProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) error {
	if err := state.CheckExistsState(statecurrency.StateKeyAccount(ipp.item.Approved()), getStateFunc); err != nil {
		return errors.Errorf("approved not found, %q: %w", ipp.item.Approved(), err)
	}

	st, err := state.ExistsState(statenft.NFTStateKey(ipp.item.contract, ipp.item.collection, statenft.CollectionKey), "key of design", getStateFunc)
	if err != nil {
		return errors.Errorf("collection design not found, %q: %w", ipp.item.contract, err)
	}

	design, err := statenft.StateCollectionValue(st)
	if err != nil {
		return errors.Errorf("collection design value not found, %q: %w", ipp.item.collection, err)
	}

	if !design.Active() {
		return errors.Errorf("deactivated collection, %q", ipp.item.collection)
	}

	st, err = state.ExistsState(stateextension.StateKeyContractAccount(design.Parent()), "contract account", getStateFunc)
	if err != nil {
		return errors.Errorf("parent not found, %q: %w", design.Parent(), err)
	}

	ca, err := stateextension.StateContractAccountValue(st)
	if err != nil {
		return errors.Errorf("contract account value not found, %q: %w", design.Parent(), err)
	}

	if !ca.IsActive() {
		return errors.Errorf("deactivated contract account, %q", design.Parent())
	}

	st, err = state.ExistsState(statenft.StateKeyNFT(ipp.item.contract, ipp.item.collection, ipp.item.idx), "key of nft", getStateFunc)
	if err != nil {
		return errors.Errorf("nft not found, %q: %w", ipp.item.idx, err)
	}

	nv, err := statenft.StateNFTValue(st)
	if err != nil {
		return errors.Errorf("nft value not found, %q: %w", ipp.item.idx, err)
	}

	if !nv.Active() {
		return errors.Errorf("burned nft, %q", ipp.item.idx)
	}

	if ipp.item.Approved().Equal(nv.Approved()) {
		return errors.Errorf("already approved, %q", ipp.item.Approved())
	}

	if ipp.item.Approved().Equal(nv.Owner()) {
		return errors.Errorf("approved account is same with owner, %q", ipp.item.Approved())
	}

	if !nv.Owner().Equal(ipp.sender) {
		if err := state.CheckExistsState(statecurrency.StateKeyAccount(nv.Owner()), getStateFunc); err != nil {
			return errors.Errorf("nft owner not found, %q: %w", nv.Owner(), err)
		}

		st, err = state.ExistsState(statenft.StateKeyOperators(ipp.item.contract, ipp.item.collection, nv.Owner()), "key of operators", getStateFunc)
		if err != nil {
			return errors.Errorf("unauthorized sender, %q: %w", ipp.sender, err)
		}

		operators, err := statenft.StateOperatorsBookValue(st)
		if err != nil {
			return errors.Errorf("operators book value not found, %q: %w", statenft.StateKeyOperators(ipp.item.contract, ipp.item.collection, nv.Owner()), err)
		}

		if !operators.Exists(ipp.sender) {
			return errors.Errorf("unauthorized sender, %q", ipp.sender)
		}
	}

	return nil
}

func (ipp *ApproveItemProcessor) Process(
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

	n := types.NewNFT(nv.ID(), nv.Active(), nv.Owner(), nv.NFTHash(), nv.URI(), ipp.item.Approved(), nv.Creators())
	if err := n.IsValid(nil); err != nil {
		return nil, err
	}

	sts := []mitumbase.StateMergeValue{state.NewStateMergeValue(st.Key(), statenft.NewNFTStateValue(n))}

	return sts, nil
}

func (ipp *ApproveItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = ApproveItem{}

	approveItemProcessorPool.Put(ipp)

	return nil
}

type ApproveProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewApproveProcessor() currencytypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new ApproveProcessor")

		nopp := approveProcessorPool.Get()
		opp, ok := nopp.(*ApproveProcessor)
		if !ok {
			return nil, e.Errorf("expected ApproveProcessor, not %T", nopp)
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

func (opp *ApproveProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess Approve")

	fact, ok := op.Fact().(ApproveFact)
	if !ok {
		return ctx, nil, e.Errorf("expected ApproveFact, not %T", op.Fact())
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	for _, item := range fact.Items() {
		ip := approveItemProcessorPool.Get()
		ipc, ok := ip.(*ApproveItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected ApproveItemProcessor, not %T", ipc)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess ApproveItem: %w", err), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *ApproveProcessor) Process(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process Approve")

	fact, ok := op.Fact().(ApproveFact)
	if !ok {
		return nil, nil, e.Errorf("expected ApproveFact, not %T", op.Fact())
	}

	var sts []mitumbase.StateMergeValue // nolint:prealloc
	for _, item := range fact.Items() {
		ip := approveItemProcessorPool.Get()
		ipc, ok := ip.(*ApproveItemProcessor)
		if !ok {
			return nil, nil, e.Errorf("expected ApproveItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = item

		s, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process ApproveItem: %w", err), nil
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
		sts = append(sts, currencystate.NewStateMergeValue(sb[i].Key(), stv))
	}

	return sts, nil, nil
}

func (opp *ApproveProcessor) Close() error {
	approveProcessorPool.Put(opp)

	return nil
}

func CalculateCollectionItemsFee(getStateFunc mitumbase.GetStateFunc, items []CollectionItem) (map[currencytypes.CurrencyID][2]common.Big, error) {
	required := map[currencytypes.CurrencyID][2]common.Big{}

	for _, item := range items {
		rq := [2]common.Big{common.ZeroBig, common.ZeroBig}

		if k, found := required[item.Currency()]; found {
			rq = k
		}

		policy, err := state.ExistsCurrencyPolicy(item.Currency(), getStateFunc)
		if err != nil {
			return nil, err
		}

		switch k, err := policy.Feeer().Fee(common.ZeroBig); {
		case err != nil:
			return nil, err
		case !k.OverZero():
			required[item.Currency()] = [2]common.Big{rq[0], rq[1]}
		default:
			required[item.Currency()] = [2]common.Big{rq[0].Add(k), rq[1].Add(k)}
		}

	}

	return required, nil

}
