package digest

import (
	"context"
	"fmt"
	"sync"
	"time"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-nft/nft/collection"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/tree"
	"github.com/spikeekips/mitum/util/valuehash"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bulkWriteLimit = 500

type BlockSession struct {
	sync.RWMutex
	block                 block.Block
	st                    *Database
	opsTreeNodes          map[string]operation.FixedTreeNode
	operationModels       []mongo.WriteModel
	accountModels         []mongo.WriteModel
	balanceModels         []mongo.WriteModel
	contractAccountModels []mongo.WriteModel
	nftCollectionModels   []mongo.WriteModel
	nftModels             []mongo.WriteModel
	nftAgentModels        []mongo.WriteModel
	statesValue           *sync.Map
	nftList               []string
}

func NewBlockSession(st *Database, blk block.Block) (*BlockSession, error) {
	if st.Readonly() {
		return nil, errors.Errorf("readonly mode")
	}

	nst, err := st.New()
	if err != nil {
		return nil, err
	}

	return &BlockSession{
		st:          nst,
		block:       blk,
		statesValue: &sync.Map{},
	}, nil
}

func (bs *BlockSession) Prepare() error {
	bs.Lock()
	defer bs.Unlock()

	if err := bs.prepareOperationsTree(); err != nil {
		return err
	}

	if err := bs.prepareOperations(); err != nil {
		return err
	}

	return bs.prepareAccounts()
}

func (bs *BlockSession) Commit(ctx context.Context) error {
	bs.Lock()
	defer bs.Unlock()

	started := time.Now()
	defer func() {
		bs.statesValue.Store("commit", time.Since(started))

		_ = bs.close()
	}()

	if err := bs.writeModels(ctx, defaultColNameOperation, bs.operationModels); err != nil {
		return err
	}

	if err := bs.writeModels(ctx, defaultColNameAccount, bs.accountModels); err != nil {
		return err
	}

	if err := bs.writeModels(ctx, defaultColNameBalance, bs.balanceModels); err != nil {
		return err
	}

	if len(bs.contractAccountModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameExtension, bs.contractAccountModels); err != nil {
			return err
		}
	}

	if len(bs.nftCollectionModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameNFTCollection, bs.nftCollectionModels); err != nil {
			return err
		}
	}

	if len(bs.nftModels) > 0 {
		for i := range bs.nftList {
			err := bs.st.cleanByHeightColNameNFTId(
				ctx,
				bs.block.Height(),
				defaultColNameNFT,
				bs.nftList[i],
			)
			if err != nil {
				return err
			}
		}

		if len(bs.nftModels) > 0 {
			if err := bs.writeModels(ctx, defaultColNameNFT, bs.nftModels); err != nil {
				return err
			}
		}
	}

	if len(bs.nftAgentModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameNFTAgent, bs.nftAgentModels); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockSession) Close() error {
	bs.Lock()
	defer bs.Unlock()

	return bs.close()
}

func (bs *BlockSession) prepareOperationsTree() error {
	nodes := map[string]operation.FixedTreeNode{}
	if err := bs.block.OperationsTree().Traverse(func(no tree.FixedTreeNode) (bool, error) {
		nno := no.(operation.FixedTreeNode)
		fh := valuehash.NewBytes(nno.Key())
		nodes[fh.String()] = nno

		return true, nil
	}); err != nil {
		return err
	}

	bs.opsTreeNodes = nodes

	return nil
}

func (bs *BlockSession) prepareOperations() error {
	if len(bs.block.Operations()) < 1 {
		return nil
	}

	node := func(h valuehash.Hash) (bool /* found */, bool /* instate */, operation.ReasonError) {
		no, found := bs.opsTreeNodes[h.String()]
		if !found {
			return false, false, nil
		}

		return true, no.InState(), no.Reason()
	}

	bs.operationModels = make([]mongo.WriteModel, len(bs.block.Operations()))

	for i := range bs.block.Operations() {
		op := bs.block.Operations()[i]

		found, inState, reason := node(op.Fact().Hash())
		if !found {
			return util.NotFoundError.Errorf("operation, %s not found in operations tree", op.Fact().Hash().String())
		}

		doc, err := NewOperationDoc(
			op,
			bs.st.database.Encoder(),
			bs.block.Height(),
			bs.block.ConfirmedAt(),
			inState,
			reason,
			uint64(i),
		)
		if err != nil {
			return err
		}
		bs.operationModels[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	return nil
}

func (bs *BlockSession) prepareAccounts() error {
	if len(bs.block.States()) < 1 {
		return nil
	}

	var accountModels []mongo.WriteModel
	var balanceModels []mongo.WriteModel
	var contractAccountModels []mongo.WriteModel
	var nftCollectionModels []mongo.WriteModel
	var nftModels []mongo.WriteModel
	var nftAgentModels []mongo.WriteModel
	for i := range bs.block.States() {
		st := bs.block.States()[i]
		switch {
		case currency.IsStateAccountKey(st.Key()):
			j, err := bs.handleAccountState(st)
			if err != nil {
				return err
			}
			accountModels = append(accountModels, j...)
		case currency.IsStateBalanceKey(st.Key()):
			j, err := bs.handleBalanceState(st)
			if err != nil {
				return err
			}
			balanceModels = append(balanceModels, j...)
		case extensioncurrency.IsStateContractAccountKey(st.Key()):
			j, err := bs.handleContractAccountState(st)
			if err != nil {
				return err
			}
			contractAccountModels = append(contractAccountModels, j...)
		case collection.IsStateCollectionKey(st.Key()):
			j, err := bs.handleNFTCollectionState(st)
			if err != nil {
				return err
			}
			nftCollectionModels = append(nftCollectionModels, j...)
		case collection.IsStateNFTKey(st.Key()):
			j, err := bs.handleNFTState(st)
			if err != nil {
				return err
			}
			nftModels = append(nftModels, j...)
		case collection.IsStateAgentKey(st.Key()):
			j, err := bs.handleNFTAgentState(st)
			if err != nil {
				return err
			}
			nftAgentModels = append(nftAgentModels, j...)
		default:
			continue
		}
	}

	bs.accountModels = accountModels
	bs.balanceModels = balanceModels

	if len(contractAccountModels) > 0 {
		bs.contractAccountModels = contractAccountModels
	}
	if len(nftCollectionModels) > 0 {
		bs.nftCollectionModels = nftCollectionModels
	}
	if len(nftModels) > 0 {
		bs.nftModels = nftModels
	}

	if len(nftAgentModels) > 0 {
		bs.nftAgentModels = nftAgentModels
	}

	return nil
}

func (bs *BlockSession) handleAccountState(st state.State) ([]mongo.WriteModel, error) {
	if rs, err := NewAccountValue(st); err != nil {
		return nil, err
	} else if doc, err := NewAccountDoc(rs, bs.st.database.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
	}
}

func (bs *BlockSession) handleBalanceState(st state.State) ([]mongo.WriteModel, error) {
	doc, err := NewBalanceDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) handleContractAccountState(st state.State) ([]mongo.WriteModel, error) {
	doc, err := NewContractAccountDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) handleNFTCollectionState(st state.State) ([]mongo.WriteModel, error) {
	doc, err := NewNFTCollectionDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}

	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) handleNFTState(st state.State) ([]mongo.WriteModel, error) {
	doc, err := NewNFTDoc(st, bs.st.database.Encoder(), bs.block.Height())
	if err != nil {
		return nil, err
	}

	bs.nftList = append(bs.nftList, doc.va.nft.ID().String())
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) handleNFTAgentState(st state.State) ([]mongo.WriteModel, error) {
	doc, err := NewNFTAgentDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}

	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) writeModels(ctx context.Context, col string, models []mongo.WriteModel) error {
	started := time.Now()
	defer func() {
		bs.statesValue.Store(fmt.Sprintf("write-models-%s", col), time.Since(started))
	}()

	n := len(models)
	if n < 1 {
		return nil
	} else if n <= bulkWriteLimit {
		return bs.writeModelsChunk(ctx, col, models)
	}

	z := n / bulkWriteLimit
	if n%bulkWriteLimit != 0 {
		z++
	}

	for i := 0; i < z; i++ {
		s := i * bulkWriteLimit
		e := s + bulkWriteLimit
		if e > n {
			e = n
		}

		if err := bs.writeModelsChunk(ctx, col, models[s:e]); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockSession) writeModelsChunk(ctx context.Context, col string, models []mongo.WriteModel) error {
	opts := options.BulkWrite().SetOrdered(false)
	if res, err := bs.st.database.Client().Collection(col).BulkWrite(ctx, models, opts); err != nil {
		return storage.MergeStorageError(err)
	} else if res != nil && res.InsertedCount < 1 {
		return errors.Errorf("not inserted to %s", col)
	}

	return nil
}

func (bs *BlockSession) close() error {
	bs.block = nil
	bs.operationModels = nil
	bs.accountModels = nil
	bs.balanceModels = nil
	bs.contractAccountModels = nil
	bs.nftCollectionModels = nil
	bs.nftModels = nil
	bs.nftAgentModels = nil

	return bs.st.Close()
}
