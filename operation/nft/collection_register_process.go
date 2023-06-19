package nft

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	statenft "github.com/ProtoconNet/mitum-nft/v2/state"
	"github.com/ProtoconNet/mitum-nft/v2/types"

	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var collectionRegisterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CollectionRegisterProcessor)
	},
}

func (CollectionRegister) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type CollectionRegisterProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewCollectionRegisterProcessor() currencytypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new CollectionRegisterProcessor")

		nopp := collectionRegisterProcessorPool.Get()
		opp, ok := nopp.(*CollectionRegisterProcessor)
		if !ok {
			return nil, errors.Errorf("expected CollectionRegisterProcessor, not %T", nopp)
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

func (opp *CollectionRegisterProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess CollectionRegister")

	fact, ok := op.Fact().(CollectionRegisterFact)
	if !ok {
		return ctx, nil, e.Errorf("expected CollectionRegisterFact, not %T", op.Fact())
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := state.CheckNotExistsState(stateextension.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender is contract account, %q", fact.Sender()), nil
	}

	if err := state.CheckFactSignsByState(fact.Sender(), op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	st, err := state.ExistsState(stateextension.StateKeyContractAccount(fact.Contract()), "key of contract account", getStateFunc)
	if err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("target contract account not found, %q: %w", fact.Contract(), err), nil
	}

	ca, err := stateextension.StateContractAccountValue(st)
	if err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("failed to get state value of contract account, %q: %w", fact.Contract(), err), nil
	}

	if !ca.Owner().Equal(fact.Sender()) {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender is not owner of contract account, %q, %q", fact.Sender(), ca.Owner()), nil
	}

	if !ca.IsActive() {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("deactivated contract account, %q", fact.Contract()), nil
	}

	if err := state.CheckNotExistsState(statenft.NFTStateKey(fact.contract, fact.Collection(), statenft.CollectionKey), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("collection design already exists, %q: %w", fact.Collection(), err), nil
	}

	if err := state.CheckNotExistsState(statenft.NFTStateKey(fact.contract, fact.Collection(), statenft.LastIDXKey), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("last index of collection design already exists, %q: %w", fact.Collection(), err), nil
	}

	whites := fact.Whites()
	for _, white := range whites {
		if err := state.CheckExistsState(statecurrency.StateKeyAccount(white), getStateFunc); err != nil {
			return ctx, mitumbase.NewBaseOperationProcessReasonError("whitelist account not found, %q: %w", white, err), nil
		} else if err = state.CheckNotExistsState(stateextension.StateKeyContractAccount(white), getStateFunc); err != nil {
			return ctx, mitumbase.NewBaseOperationProcessReasonError("whitelist account is contract account, %q: %w", white, err), nil
		}
	}

	return ctx, nil, nil
}

func (opp *CollectionRegisterProcessor) Process(
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process CollectionRegister")

	fact, ok := op.Fact().(CollectionRegisterFact)
	if !ok {
		return nil, nil, e.Errorf("expected CollectionRegisterFact, not %T", op.Fact())
	}

	sts := make([]mitumbase.StateMergeValue, 3)

	policy := types.NewCollectionPolicy(fact.Name(), fact.Royalty(), fact.URI(), fact.Whites())
	design := types.NewDesign(fact.Contract(), fact.Sender(), fact.Collection(), true, policy)
	if err := design.IsValid(nil); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("invalid collection design, %q: %w", fact.Collection(), err), nil
	}

	sts[0] = currencystate.NewStateMergeValue(
		statenft.NFTStateKey(design.Parent(), design.Collection(), statenft.CollectionKey),
		statenft.NewCollectionStateValue(design),
	)
	sts[1] = currencystate.NewStateMergeValue(
		statenft.NFTStateKey(design.Parent(), design.Collection(), statenft.LastIDXKey),
		statenft.NewLastNFTIndexStateValue(0),
	)

	currencyPolicy, err := state.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("currency not found, %q: %w", fact.Currency(), err), nil
	}

	fee, err := currencyPolicy.Feeer().Fee(common.ZeroBig)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check fee of currency, %q: %w", fact.Currency(), err), nil
	}

	st, err := state.ExistsState(statecurrency.StateKeyBalance(fact.Sender(), fact.Currency()), "key of sender balance", getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("sender balance not found, %q: %w", fact.Sender(), err), nil
	}
	sb := currencystate.NewStateMergeValue(st.Key(), st.Value())

	switch b, err := statecurrency.StateBalanceValue(st); {
	case err != nil:
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to get balance value, %q: %w", statecurrency.StateKeyBalance(fact.Sender(), fact.Currency()), err), nil
	case b.Big().Compare(fee) < 0:
		return nil, mitumbase.NewBaseOperationProcessReasonError("not enough balance of sender, %q", fact.Sender()), nil
	}

	v, ok := sb.Value().(statecurrency.BalanceStateValue)
	if !ok {
		return nil, mitumbase.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", sb.Value()), nil
	}
	sts[2] = currencystate.NewStateMergeValue(
		sb.Key(),
		statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(fee))),
	)

	return sts, nil, nil
}

func (opp *CollectionRegisterProcessor) Close() error {
	collectionRegisterProcessorPool.Put(opp)

	return nil
}
