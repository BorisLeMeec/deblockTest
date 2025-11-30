package service

import (
	"context"
	"log"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"deblockTest/pkg"
)

type Worker struct {
	client     EthereumBlockGetter
	userGetter UserGetter
	publisher  Publisher
	state      State
	blocks     <-chan uint64
	processed  *atomic.Uint64
	retryChan  chan<- uint64
	ackChan    chan<- uint64
}

func (w *Worker) Run(ctx context.Context) {
	for blockNum := range w.blocks {
		block, err := w.client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
		if err != nil {
			log.Printf("Failed to fetch block %d: %v (will retry later)", blockNum, err)
			time.Sleep(100 * time.Millisecond)
			// This retry mechanism will break the in-order processing, but acceptable in 99% of case.
			// if not acceptable we can introduce a local retry mechanism to make sure we handle each block after the previous one.
			go func(b uint64) { w.retryChan <- b }(blockNum)
			continue
		}

		msgs := w.processBlock(block)
		if len(msgs) > 0 {
			w.publisher.Publish(ctx, msgs)
		}
		w.ackChan <- block.NumberU64()
	}
}

func (w *Worker) processBlock(block *types.Block) []pkg.TxMessage {
	var msgs []pkg.TxMessage
	// log.Printf("Processing block %d\n", block.NumberU64())

	for _, tx := range block.Transactions() {
		chainID := tx.ChainId()
		if chainID.Sign() == 0 {
			continue
		}
		signer := types.NewLondonSigner(chainID)
		from, _ := types.Sender(signer, tx)
		to := tx.To()

		userFrom, hasFrom := w.userGetter.GetUserID(from)
		userTo := ""
		hasTo := false
		if to != nil {
			userTo, hasTo = w.userGetter.GetUserID(*to)
		}

		if !hasFrom && !hasTo {
			continue
		}

		if hasFrom {
			msgs = append(msgs, pkg.TxMessage{
				UserID:      userFrom,
				From:        from.Hex(),
				To:          pkg.ToStringPtr(to),
				Amount:      tx.Value().String(),
				Hash:        tx.Hash().Hex(),
				BlockNumber: block.NumberU64(),
			})
		}

		if hasTo {
			msgs = append(msgs, pkg.TxMessage{
				UserID:      userTo,
				From:        from.Hex(),
				To:          to.Hex(),
				Amount:      tx.Value().String(),
				Hash:        tx.Hash().Hex(),
				BlockNumber: block.NumberU64(),
			})
		}
	}
	// log.Printf("Processed %d transactions, produced %d messages\n", len(block.Transactions()), len(msgs))
	return msgs
}
