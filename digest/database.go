package digest

import (
	"context"
	"math"
	"sort"
	"strconv"
	"sync"

	"github.com/ProtoconNet/mitum-nft/v2/state"
	"github.com/ProtoconNet/mitum-nft/v2/types"

	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mongodbstorage "github.com/ProtoconNet/mitum-nft/v2/digest/mongodb"
	"github.com/ProtoconNet/mitum-nft/v2/digest/util"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var maxLimit int64 = 50

var (
	defaultColNameAccount       = "digest_ac"
	defaultColNameBalance       = "digest_bl"
	defaultColNameCurrency      = "digest_cr"
	defaultColNameOperation     = "digest_op"
	defaultColNameTimeStamp     = "digest_ts"
	defaultColNameBlock         = "digest_bm"
	defaultColNameNFTCollection = "digest_nftcollection"
	defaultColNameNFT           = "digest_nft"
	defaultColNameNFTOperator   = "digest_nftoperator"
)

var AllCollections = []string{
	defaultColNameAccount,
	defaultColNameBalance,
	defaultColNameCurrency,
	defaultColNameOperation,
	defaultColNameTimeStamp,
	defaultColNameBlock,
	defaultColNameNFTCollection,
	defaultColNameNFT,
	defaultColNameNFTOperator,
}

var DigestStorageLastBlockKey = "digest_last_block"

type Database struct {
	sync.RWMutex
	*logging.Logging
	mitum     *isaacdatabase.Center
	database  *mongodbstorage.Database
	readonly  bool
	lastBlock mitumbase.Height
}

func NewDatabase(mitum *isaacdatabase.Center, st *mongodbstorage.Database) (*Database, error) {
	nst := &Database{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "digest-mongodb-database")
		}),
		mitum:     mitum,
		database:  st,
		lastBlock: mitumbase.NilHeight,
	}
	_ = nst.SetLogging(mitum.Logging)

	return nst, nil
}

func NewReadonlyDatabase(mitum *isaacdatabase.Center, st *mongodbstorage.Database) (*Database, error) {
	nst, err := NewDatabase(mitum, st)
	if err != nil {
		return nil, err
	}
	nst.readonly = true

	return nst, nil
}

func (st *Database) New() (*Database, error) {
	if st.readonly {
		return nil, errors.Errorf("readonly mode")
	}

	nst, err := st.database.New()
	if err != nil {
		return nil, err
	}
	return NewDatabase(st.mitum, nst)
}

func (st *Database) Readonly() bool {
	return st.readonly
}

func (st *Database) Close() error {
	return st.database.Close()
}

func (st *Database) Initialize() error {
	st.Lock()
	defer st.Unlock()

	switch h, found, err := loadLastBlock(st); {
	case err != nil:
		return errors.Wrap(err, "failed to initialize digest database")
	case !found:
		st.lastBlock = mitumbase.NilHeight
		st.Log().Debug().Msg("last block for digest not found")
	default:
		st.lastBlock = h
	}
	// 	if !st.readonly {
	// 		if err := st.createIndex(); err != nil {
	// 			return err
	// 		}

	// 		if err := st.cleanByHeight(context.Background(), h+1); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	if !st.readonly {
		if err := st.createIndex(); err != nil {
			return err
		}

		// if err := st.cleanByHeight(context.Background(), h+1); err != nil {
		// 	return err
		// }
	}

	return nil
}

func (st *Database) createIndex() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	for col, models := range defaultIndexes {
		if err := st.database.CreateIndex(col, models, indexPrefix); err != nil {
			return err
		}
	}

	return nil
}

func (st *Database) LastBlock() mitumbase.Height {
	st.RLock()
	defer st.RUnlock()

	return st.lastBlock
}

func (st *Database) SetLastBlock(height mitumbase.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	if height <= st.lastBlock {
		return nil
	}

	return st.setLastBlock(height)
}

func (st *Database) setLastBlock(height mitumbase.Height) error {
	if err := st.database.SetInfo(DigestStorageLastBlockKey, height.Bytes()); err != nil {
		st.Log().Debug().Int64("height", height.Int64()).Msg("failed to set last block")

		return err
	}
	st.lastBlock = height
	st.Log().Debug().Int64("height", height.Int64()).Msg("set last block")

	return nil
}

func (st *Database) Clean() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	return st.clean(context.Background())
}

func (st *Database) clean(ctx context.Context) error {
	for _, col := range []string{
		defaultColNameAccount,
		defaultColNameBalance,
		defaultColNameCurrency,
		defaultColNameOperation,
		defaultColNameTimeStamp,
		defaultColNameBlock,
		defaultColNameNFTCollection,
		defaultColNameNFT,
		defaultColNameNFTOperator,
	} {
		if err := st.database.Client().Collection(col).Drop(ctx); err != nil {
			return err
		}

		st.Log().Debug().Str("collection", col).Msg("drop collection by height")
	}

	if err := st.setLastBlock(mitumbase.NilHeight); err != nil {
		return err
	}

	st.Log().Debug().Msg("clean digest")

	return nil
}

func (st *Database) CleanByHeight(ctx context.Context, height mitumbase.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	return st.cleanByHeight(ctx, height)
}

func (st *Database) cleanByHeight(ctx context.Context, height mitumbase.Height) error {
	if height <= mitumbase.GenesisHeight {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByHeight := mongo.NewDeleteManyModel().SetFilter(bson.M{"height": bson.M{"$gte": height}})

	for _, col := range []string{
		defaultColNameAccount,
		defaultColNameBalance,
		defaultColNameCurrency,
		defaultColNameOperation,
		defaultColNameTimeStamp,
		defaultColNameBlock,
		defaultColNameNFTCollection,
		defaultColNameNFT,
		defaultColNameNFTOperator,
	} {
		res, err := st.database.Client().Collection(col).BulkWrite(
			ctx,
			[]mongo.WriteModel{removeByHeight},
			opts,
		)
		if err != nil {
			return err
		}

		st.Log().Debug().Str("collection", col).Interface("result", res).Msg("clean collection by height")
	}

	return st.setLastBlock(height - 1)
}

/*
func (st *Database) Manifest(h mitumutil.Hash) (mitumbase.Manifest, bool, error) {
	return st.mitum.Manifest(h)
}
*/

// Manifests returns block.Manifests by it's order, height.
func (st *Database) Manifests(
	load bool,
	reverse bool,
	offset mitumbase.Height,
	limit int64,
	callback func(mitumbase.Height, mitumbase.Manifest, uint64) (bool, error),
) error {
	var filter bson.M
	if offset > mitumbase.NilHeight {
		if reverse {
			filter = bson.M{"height": bson.M{"$lt": offset}}
		} else {
			filter = bson.M{"height": bson.M{"$gt": offset}}
		}
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameBlock,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			va, ops, err := LoadManifest(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Height(), va, ops)

		},
		opt,
	)
}

// OperationsByAddress finds the operation.Operations, which are related with
// the given Address. The returned valuehash.Hash is the
// operation.Operation.Fact().Hash().
// *    load:if true, load operation.Operation and returns it. If not, just hash will be returned
// * reverse: order by height; if true, higher height will be returned first.
// *  offset: returns from next of offset, usually it is combination of
// "<height>,<fact>".
func (st *Database) OperationsByAddress(
	address mitumbase.Address,
	load,
	reverse bool,
	offset string,
	limit int64,
	callback func(mitumutil.Hash /* fact hash */, OperationValue) (bool, error),
) error {
	filter, err := buildOperationsFilterByAddress(address, offset, reverse)
	if err != nil {
		return err
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	if !load {
		opt = opt.SetProjection(bson.M{"fact": 1})
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameOperation,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if !load {
				h, err := LoadOperationHash(cursor.Decode)
				if err != nil {
					return false, err
				}
				return callback(h, OperationValue{})
			}

			va, err := LoadOperation(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Operation().Fact().Hash(), va)
		},
		opt,
	)
}

// Operation returns operation.Operation. If load is false, just returns nil
// Operation.
func (st *Database) Operation(
	h mitumutil.Hash, /* fact hash */
	load bool,
) (OperationValue, bool /* exists */, error) {
	if !load {
		exists, err := st.database.Client().Exists(defaultColNameOperation, util.NewBSONFilter("fact", h).D())
		return OperationValue{}, exists, err
	}

	var va OperationValue
	if err := st.database.Client().GetByFilter(
		defaultColNameOperation,
		util.NewBSONFilter("fact", h).D(),
		func(res *mongo.SingleResult) error {
			if !load {
				return nil
			}

			i, err := LoadOperation(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			va = i

			return nil
		},
	); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return OperationValue{}, false, nil
		}

		return OperationValue{}, false, err
	}
	return va, true, nil
}

// Operations returns operation.Operations by it's order, height and index.
func (st *Database) Operations(
	filter bson.M,
	load bool,
	reverse bool,
	limit int64,
	callback func(mitumutil.Hash /* fact hash */, OperationValue, int64) (bool, error),
) error {
	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	if !load {
		opt = opt.SetProjection(bson.M{"fact": 1})
	}

	count, err := st.database.Client().Count(context.Background(), defaultColNameOperation, bson.D{})
	if err != nil {
		return err
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameOperation,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if !load {
				h, err := LoadOperationHash(cursor.Decode)
				if err != nil {
					return false, err
				}
				return callback(h, OperationValue{}, count)
			}

			va, err := LoadOperation(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Operation().Fact().Hash(), va, count)
		},
		opt,
	)
}

// Account returns AccountValue.
func (st *Database) Account(a mitumbase.Address) (AccountValue, bool /* exists */, error) {
	var rs AccountValue
	if err := st.database.Client().GetByFilter(
		defaultColNameAccount,
		util.NewBSONFilter("address", a.String()).D(),
		func(res *mongo.SingleResult) error {
			i, err := LoadAccountValue(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			rs = i

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		if errors.Is(err, mitumutil.NewIDError("not found")) {
			return rs, false, nil
		}

		return rs, false, err
	}

	// NOTE load balance
	switch am, lastHeight, err := st.balance(a); {
	case err != nil:
		return rs, false, err
	default:
		rs = rs.SetBalance(am).
			SetHeight(lastHeight)
	}

	return rs, true, nil
}

// AccountsByPublickey finds Accounts, which are related with the given
// Publickey.
// *  offset: returns from next of offset, usually it is "<height>,<address>".
func (st *Database) AccountsByPublickey(
	pub mitumbase.Publickey,
	loadBalance bool,
	offsetHeight mitumbase.Height,
	offsetAddress string,
	limit int64,
	callback func(AccountValue) (bool, error),
) error {
	if offsetHeight <= mitumbase.NilHeight {
		return errors.Errorf("offset height should be over nil height")
	}

	filter := buildAccountsFilterByPublickey(pub)
	filter["height"] = bson.M{"$lte": offsetHeight}

	var sas []string
	switch i, err := st.addressesByPublickey(filter); {
	case err != nil:
		return err
	default:
		sas = i
	}

	if len(sas) < 1 {
		return nil
	}

	var filteredAddress []string
	if len(offsetAddress) < 1 {
		filteredAddress = sas
	} else {
		var found bool
		for i := range sas {
			a := sas[i]
			if !found {
				if offsetAddress == a {
					found = true
				}

				continue
			}

			filteredAddress = append(filteredAddress, a)
		}
	}

	if len(filteredAddress) < 1 {
		return nil
	}

end:
	for i := int64(0); i < int64(math.Ceil(float64(len(filteredAddress))/50.0)); i++ {
		l := (i + 1) + 50
		if n := int64(len(filteredAddress)); l > n {
			l = n
		}

		limited := filteredAddress[i*50 : l]
		switch done, err := st.filterAccountByPublickey(
			pub, limited, limit, loadBalance, callback,
		); {
		case err != nil:
			return err
		case done:
			break end
		}
	}

	return nil
}

func (st *Database) balance(a mitumbase.Address) ([]currencytypes.Amount, mitumbase.Height, error) {
	lastHeight := mitumbase.NilHeight
	var cids []string

	amm := map[currencytypes.CurrencyID]currencytypes.Amount{}
	for {
		filter := util.NewBSONFilter("address", a.String())

		var q primitive.D
		if len(cids) < 1 {
			q = filter.D()
		} else {
			q = filter.Add("currency", bson.M{"$nin": cids}).D()
		}

		var sta mitumbase.State
		if err := st.database.Client().GetByFilter(
			defaultColNameBalance,
			q,
			func(res *mongo.SingleResult) error {
				i, err := LoadBalance(res.Decode, st.database.Encoders())
				if err != nil {
					return err
				}
				sta = i

				return nil
			},
			options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
		); err != nil {
			if err.Error() == mitumutil.NewIDError("mongo: no documents in result").Error() {
				break
			}

			return nil, lastHeight, err
		}

		i, err := statecurrency.StateBalanceValue(sta)
		if err != nil {
			return nil, lastHeight, err
		}
		amm[i.Currency()] = i

		cids = append(cids, i.Currency().String())

		if h := sta.Height(); h > lastHeight {
			lastHeight = h
		}
	}

	ams := make([]currencytypes.Amount, len(amm))
	var i int
	for k := range amm {
		ams[i] = amm[k]
		i++
	}

	return ams, lastHeight, nil
}

func (st *Database) currencies() ([]string, error) {
	var cids []string

	for {
		filter := util.EmptyBSONFilter()

		var q primitive.D
		if len(cids) < 1 {
			q = filter.D()
		} else {
			q = filter.Add("currency", bson.M{"$nin": cids}).D()
		}

		opt := options.FindOne().SetSort(
			util.NewBSONFilter("height", -1).D(),
		)
		var sta mitumbase.State
		if err := st.database.Client().GetByFilter(
			defaultColNameCurrency,
			q,
			func(res *mongo.SingleResult) error {
				i, err := LoadState(res.Decode, st.database.Encoders())
				if err != nil {
					return err
				}
				sta = i
				return nil
			},
			opt,
		); err != nil {
			if err.Error() == mitumutil.NewIDError("mongo: no documents in result").Error() {
				break
			}

			return nil, err
		}

		if sta != nil {
			i, err := statecurrency.StateCurrencyDesignValue(sta)
			if err != nil {
				return nil, err
			}
			cids = append(cids, i.Currency().String())
		} else {
			return nil, errors.Errorf("state is nil")
		}

	}

	return cids, nil
}

func (st *Database) ManifestByHeight(height mitumbase.Height) (mitumbase.Manifest, uint64, error) {
	q := util.NewBSONFilter("height", height).D()

	var m mitumbase.Manifest
	var operations uint64
	if err := st.database.Client().GetByFilter(
		defaultColNameBlock,
		q,
		func(res *mongo.SingleResult) error {
			v, ops, err := LoadManifest(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			m = v
			operations = ops
			return nil
		},
	); err != nil {
		return nil, 0, err
	}

	if m != nil {
		return m, operations, nil
	} else {
		return nil, 0, errors.Errorf("manifest is nil")
	}
}

func (st *Database) ManifestByHash(hash mitumutil.Hash) (mitumbase.Manifest, uint64, error) {
	q := util.NewBSONFilter("block", hash).D()

	var m mitumbase.Manifest
	var operations uint64
	if err := st.database.Client().GetByFilter(
		defaultColNameBlock,
		q,
		func(res *mongo.SingleResult) error {
			v, ops, err := LoadManifest(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			m = v
			operations = ops
			return nil
		},
	); err != nil {
		return nil, 0, err
	}

	if m != nil {
		return m, operations, nil
	} else {
		return nil, 0, errors.Errorf("manifest is nil")
	}
}

func (st *Database) currency(cid string) (currencytypes.CurrencyDesign, mitumbase.State, error) {
	q := util.NewBSONFilter("currency", cid).D()

	opt := options.FindOne().SetSort(
		util.NewBSONFilter("height", -1).D(),
	)
	var sta mitumbase.State
	if err := st.database.Client().GetByFilter(
		defaultColNameCurrency,
		q,
		func(res *mongo.SingleResult) error {
			i, err := LoadState(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			sta = i
			return nil
		},
		opt,
	); err != nil {
		return currencytypes.CurrencyDesign{}, nil, err
	}

	if sta != nil {
		de, err := statecurrency.StateCurrencyDesignValue(sta)
		if err != nil {
			return currencytypes.CurrencyDesign{}, nil, err
		}
		return de, sta, nil
	} else {
		return currencytypes.CurrencyDesign{}, nil, errors.Errorf("state is nil")
	}
}

func (st *Database) topHeightByPublickey(pub mitumbase.Publickey) (mitumbase.Height, error) {
	var sas []string
	switch r, err := st.database.Client().Collection(defaultColNameAccount).Distinct(
		context.Background(),
		"address",
		buildAccountsFilterByPublickey(pub),
	); {
	case err != nil:
		return mitumbase.NilHeight, err
	case len(r) < 1:
		return mitumbase.NilHeight, err
	default:
		sas = make([]string, len(r))
		for i := range r {
			sas[i] = r[i].(string)
		}
	}

	var top mitumbase.Height
	for i := int64(0); i < int64(math.Ceil(float64(len(sas))/50.0)); i++ {
		l := (i + 1) + 50
		if n := int64(len(sas)); l > n {
			l = n
		}

		switch h, err := st.partialTopHeightByPublickey(sas[i*50 : l]); {
		case err != nil:
			return mitumbase.NilHeight, err
		case top <= mitumbase.NilHeight:
			top = h
		case h > top:
			top = h
		}
	}

	return top, nil
}

func (st *Database) partialTopHeightByPublickey(as []string) (mitumbase.Height, error) {
	var top mitumbase.Height
	err := st.database.Client().Find(
		context.Background(),
		defaultColNameAccount,
		bson.M{"address": bson.M{"$in": as}},
		func(cursor *mongo.Cursor) (bool, error) {
			h, err := loadHeightDoc(cursor.Decode)
			if err != nil {
				return false, err
			}

			top = h

			return false, nil
		},
		options.Find().
			SetSort(util.NewBSONFilter("height", -1).D()).
			SetLimit(1),
	)

	return top, err
}

func (st *Database) addressesByPublickey(filter bson.M) ([]string, error) {
	r, err := st.database.Client().Collection(defaultColNameAccount).Distinct(context.Background(), "address", filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get distinct addresses")
	}

	if len(r) < 1 {
		return nil, nil
	}

	sas := make([]string, len(r))
	for i := range r {
		sas[i] = r[i].(string)
	}

	sort.Strings(sas)

	return sas, nil
}

func (st *Database) filterAccountByPublickey(
	pub mitumbase.Publickey,
	addresses []string,
	limit int64,
	loadBalance bool,
	callback func(AccountValue) (bool, error),
) (bool, error) {
	filter := bson.M{"address": bson.M{"$in": addresses}}

	var lastAddress string
	var called int64
	var stopped bool
	if err := st.database.Client().Find(
		context.Background(),
		defaultColNameAccount,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if called == limit {
				return false, nil
			}

			doc, err := loadBriefAccountDoc(cursor.Decode)
			if err != nil {
				return false, err
			}

			if len(lastAddress) > 0 {
				if lastAddress == doc.Address {
					return true, nil
				}
			}
			lastAddress = doc.Address

			if !doc.pubExists(pub) {
				return true, nil
			}

			va, err := LoadAccountValue(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}

			if loadBalance { // NOTE load balance
				switch am, lastHeight, err := st.balance(va.Account().Address()); {
				case err != nil:
					return false, err
				default:
					va = va.SetBalance(am).
						SetHeight(lastHeight)
				}
			}

			called++
			switch keep, err := callback(va); {
			case err != nil:
				return false, err
			case !keep:
				stopped = true

				return false, nil
			default:
				return true, nil
			}
		},
		options.Find().SetSort(util.NewBSONFilter("address", 1).Add("height", -1).D()),
	); err != nil {
		return false, err
	}

	return stopped || called == limit, nil
}

func (st *Database) cleanBalanceByHeightAndAccount(ctx context.Context, height mitumbase.Height, address string) error {
	if height <= mitumbase.GenesisHeight+1 {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByAddress := mongo.NewDeleteManyModel().SetFilter(bson.M{"address": address, "height": bson.M{"$lte": height}})

	res, err := st.database.Client().Collection(defaultColNameBalance).BulkWrite(
		context.Background(),
		[]mongo.WriteModel{removeByAddress},
		opts,
	)
	if err != nil {
		return err
	}

	st.Log().Debug().Str("collection", defaultColNameBalance).Interface("result", res).Msg("clean Balancecollection by address")

	return st.setLastBlock(height - 1)
}

func loadLastBlock(st *Database) (mitumbase.Height, bool, error) {
	switch b, found, err := st.database.Info(DigestStorageLastBlockKey); {
	case err != nil:
		return mitumbase.NilHeight, false, errors.Wrap(err, "failed to get last block for digest")
	case !found:
		return mitumbase.NilHeight, false, nil
	default:
		h, err := mitumbase.ParseHeightBytes(b)
		if err != nil {
			return mitumbase.NilHeight, false, err
		}
		return h, true, nil
	}
}

type heightDoc struct {
	H mitumbase.Height `bson:"height"`
}

func loadHeightDoc(decoder func(interface{}) error) (mitumbase.Height, error) {
	var h heightDoc
	if err := decoder(&h); err != nil {
		return mitumbase.NilHeight, err
	}

	return h.H, nil
}

type briefAccountDoc struct {
	ID      primitive.ObjectID `bson:"_id"`
	Address string             `bson:"address"`
	Pubs    []string           `bson:"pubs"`
	Height  mitumbase.Height   `bson:"height"`
}

func (doc briefAccountDoc) pubExists(k mitumbase.PKKey) bool {
	if len(doc.Pubs) < 1 {
		return false
	}

	for i := range doc.Pubs {
		if k.String() == doc.Pubs[i] {
			return true
		}
	}

	return false
}

func loadBriefAccountDoc(decoder func(interface{}) error) (briefAccountDoc, error) {
	var a briefAccountDoc
	if err := decoder(&a); err != nil {
		return a, err
	}

	return a, nil
}

func (st *Database) NFTCollection(contract, col string) (*types.Design, mitumbase.State, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("collection", col)

	var design *types.Design
	var sta mitumbase.State
	var err error
	if err := st.database.Client().GetByFilter(
		defaultColNameNFTCollection,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = LoadState(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}

			design, err = state.StateCollectionValue(sta)
			if err != nil {
				return err
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, nil, err
	}

	return design, nil, nil
}

func (st *Database) NFT(contract, col, idx string) (*types.NFT, mitumbase.State, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("collection", col)
	filter = filter.Add("nftid", idx)

	var nft *types.NFT
	var sta mitumbase.State
	var err error
	if err = st.database.Client().GetByFilter(
		defaultColNameNFT,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = LoadState(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			nft, err = state.StateNFTValue(sta)
			if err != nil {
				return err
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, nil, err
	}

	return nft, sta, nil
}

func (st *Database) NFTsByAddress(
	address mitumbase.Address,
	reverse bool,
	offset string,
	limit int64,
	collectionid string,
	callback func(string /* nft id */, types.NFT) (bool, error),
) error {
	filter, err := buildNFTsFilterByAddress(address, offset, reverse, collectionid)
	if err != nil {
		return err
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameNFT,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			st, err := LoadState(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			nft, err := state.StateNFTValue(st)
			if err != nil {
				return false, err
			}

			return callback(strconv.FormatUint(nft.ID(), 10), *nft)
		},
		opt,
	)
}

func (st *Database) NFTsByCollection(
	contract,
	col string,
	reverse bool,
	offset string,
	limit int64,
	callback func(nft types.NFT, st mitumbase.State) (bool, error),
) error {
	filter, err := buildNFTsFilterByCollection(contract, col, offset, reverse)
	if err != nil {
		return err
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameNFT,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			st, err := LoadState(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			nft, err := state.StateNFTValue(st)
			if err != nil {
				return false, err
			}
			return callback(*nft, st)
		},
		opt,
	)
}

func (st *Database) NFTOperators(contract, col, account string) (*types.OperatorsBook, mitumbase.State, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("collection", col)
	filter = filter.Add("account", account)

	var operators *types.OperatorsBook
	var sta mitumbase.State
	var err error
	if err := st.database.Client().GetByFilter(
		defaultColNameNFTOperator,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = LoadState(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}

			operators, err = state.StateOperatorsBookValue(sta)
			if err != nil {
				return err
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, nil, err
	}

	return operators, nil, nil
}

func (st *Database) cleanByHeightColName(
	ctx context.Context,
	height mitumbase.Height,
	colName, key, value string,
) error {
	if height <= mitumbase.GenesisHeight {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByHeight := mongo.NewDeleteManyModel().SetFilter(
		bson.M{key: value, "height": bson.M{"$lte": height}},
	)

	res, err := st.database.Client().Collection(colName).BulkWrite(
		ctx,
		[]mongo.WriteModel{removeByHeight},
		opts,
	)
	if err != nil {
		return err
	}

	st.Log().Debug().Str("collection", colName).Interface("result", res).Msg("clean collection by height")

	return st.setLastBlock(height - 1)
}
