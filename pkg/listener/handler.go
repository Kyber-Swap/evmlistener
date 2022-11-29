package listener

import (
	"bytes"
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// Handler ...
type Handler struct {
	topic string

	evmClient   EVMClient
	blockKeeper block.Keeper
	publisher   pubsub.Publisher
	l           *zap.SugaredLogger
}

// NewHandler ...
func NewHandler(
	topic string, evmClient EVMClient, blockKeeper block.Keeper, publisher pubsub.Publisher,
) *Handler {
	return &Handler{
		topic:       topic,
		evmClient:   evmClient,
		blockKeeper: blockKeeper,
		publisher:   publisher,
		l:           zap.S(),
	}
}

// Init ...
func (h *Handler) Init(ctx context.Context) error {
	err := h.blockKeeper.Init()
	if err != nil {
		h.l.Errorw("Fail to initialize block keeper", "error", err)

		return err
	}

	if h.blockKeeper.Len() > 0 {
		return nil
	}

	h.l.Info("Get latest block number")
	toBlock, err := h.evmClient.BlockNumber(ctx)
	if err != nil {
		h.l.Errorw("Fail to get latest block number", "error", err)

		return err
	}

	fromBlock := toBlock - uint64(h.blockKeeper.Cap())

	h.l.Infow("Get blocks from node", "from", fromBlock, "to", toBlock)
	blocks, err := getBlocks(ctx, h.evmClient, fromBlock, toBlock)
	if err != nil {
		h.l.Errorw("Fail to get blocks", "from", fromBlock, "to", toBlock, "error", err)

		return err
	}

	h.l.Infow("Add new blocks to block storage", "len", len(blocks))
	for _, b := range blocks {
		err = h.blockKeeper.Add(b)
		if err != nil {
			h.l.Errorw("Fail to store block", "block", b, "error", err)

			return err
		}
	}

	return nil
}

// getBlock returns block from block keeper or fetch from evm client.
func (h *Handler) getBlock(ctx context.Context, hash common.Hash) (types.Block, error) {
	b, err := h.blockKeeper.Get(hash)
	if err == nil {
		return b, nil
	}

	if !errors.Is(err, errors.ErrNotFound) {
		return types.Block{}, err
	}

	return getBlockByHash(ctx, h.evmClient, hash)
}

func (h *Handler) findReorgBlocks(
	ctx context.Context, storedBlock, newBlock types.Block,
) ([]types.Block, []types.Block, error) {
	h.l.Debugw("Find re-organization blocks", "storedBlock", storedBlock, "newBlock", newBlock)

	reorgBlocks := []types.Block{storedBlock}
	newBlocks := []types.Block{newBlock}
	storedNumber := storedBlock.Number.Uint64()
	newNumber := newBlock.Number.Uint64()

	for !bytes.Equal(storedBlock.ParentHash.Bytes(), newBlock.ParentHash.Bytes()) {
		if storedNumber >= newNumber {
			tmp, err := h.blockKeeper.Get(storedBlock.ParentHash)
			if err != nil {
				h.l.Errorw("Fail to get stored block",
					"hash", storedBlock.ParentHash, "error", err)

				return nil, nil, err
			}

			storedBlock = tmp
			storedNumber--
			reorgBlocks = append(reorgBlocks, storedBlock)
		}

		if newNumber > storedNumber {
			tmp, err := h.getBlock(ctx, newBlock.ParentHash)
			if err != nil {
				h.l.Errorw("Fail to get new block",
					"hash", newBlock.ParentHash, "error", err)

				return nil, nil, err
			}

			newBlock = tmp
			newNumber--
			newBlocks = append(newBlocks, newBlock)
		}
	}

	n := len(newBlocks)
	for i := 0; i < n/2; i++ {
		newBlocks[i], newBlocks[n-i-1] = newBlocks[n-i-1], newBlocks[i]
	}

	return reorgBlocks, newBlocks, nil
}

func (h *Handler) handleReorgBlock(
	ctx context.Context, b types.Block,
) (revertedBlocks []types.Block, newBlocks []types.Block, err error) {
	head, err := h.blockKeeper.Head()
	if err != nil {
		h.l.Errorw("Fail to get stored block head", "error", err)

		return nil, nil, err
	}

	return h.findReorgBlocks(ctx, head, b)
}

// Handle ...
func (h *Handler) Handle(ctx context.Context, b types.Block) error {
	log := h.l.With(
		"blockNumber", b.Number, "blockHash", b.Hash,
		"parentHash", b.ParentHash, "numLogs", len(b.Logs),
	)

	exists, err := h.blockKeeper.Exists(b.Hash)
	if err != nil {
		log.Errorw("Fail to check exists for block", "error", err)

		return err
	}

	if exists {
		log.Debugw("Ignore already handled block")

		return nil
	}

	isReorg, err := h.blockKeeper.IsReorg(b)
	if err != nil {
		log.Errorw("Fail to check for re-organization", "error", err)

		return err
	}

	var revertedBlocks, newBlocks []types.Block
	if isReorg {
		log.Infow("Handle re-organization block")
		revertedBlocks, newBlocks, err = h.handleReorgBlock(ctx, b)
		if err != nil {
			log.Errorw("Fail to handle re-organization block", "error", err)

			return err
		}
	} else {
		if err != nil {
			log.Errorw("Fail to store new block", "error", err)

			return err
		}

		newBlocks = []types.Block{b}
	}

	msg := types.Message{
		RevertedBlocks: revertedBlocks,
		NewBlocks:      newBlocks,
	}
	err = h.publisher.Publish(ctx, h.topic, msg)
	if err != nil {
		log.Errorw("Fail to publish message", "msg", msg, "error", err)

		return err
	}

	// Add new blocks into block keeper.
	for _, b := range newBlocks {
		err = h.blockKeeper.Add(b)
		if err != nil {
			h.l.Errorw("Fail to add block", "hash", b.Hash, "error", err)

			return err
		}
	}

	return nil
}
