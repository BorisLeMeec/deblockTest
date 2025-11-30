//go:generate mockgen -typed -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks
package service

import (
	"context"
	"log"
	"math/big"
	"time"

	"deblockTest/pkg"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthereumBlockGetter interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

type Publisher interface {
	Publish(ctx context.Context, msgs []pkg.TxMessage)
}

type UserGetter interface {
	GetUserID(addr common.Address) (string, bool)
}

type State interface {
	SaveCheckpoint(blockNum uint64) error
	LoadCheckpoint() uint64
}

type Service struct {
	config         *Config
	client         EthereumBlockGetter
	userGetter     UserGetter
	publisher      Publisher
	state          State
	blocks         chan uint64
	retryChan      chan uint64
	ackChan        chan uint64
	processedCount uint64
}

func NewService(config *Config, ethClient EthereumBlockGetter, ug UserGetter, p Publisher, s State) *Service {
	return &Service{
		config:     config,
		client:     ethClient,
		userGetter: ug,
		publisher:  p,
		state:      s,
	}
}

func (s *Service) Setup(ctx context.Context) {
	s.blocks = make(chan uint64, 1000)
	s.retryChan = make(chan uint64, 1000)
	s.ackChan = make(chan uint64, 1000)

	for i := 0; i < s.config.WorkerCount; i++ {
		w := Worker{
			client:     s.client,
			userGetter: s.userGetter,
			publisher:  s.publisher,
			blocks:     s.blocks,
			retryChan:  s.retryChan,
			ackChan:    s.ackChan,
		}
		go w.Run(ctx)
	}
}

func (s *Service) Run(ctx context.Context) {
	// Automatically merge blocks to rety in blockChan.
	go func() {
		for blockNum := range s.retryChan {
			time.Sleep(200 * time.Millisecond)
			s.blocks <- blockNum
		}
	}()

	checkpoint := s.state.LoadCheckpoint()
	latest, err := s.client.BlockNumber(ctx)
	if err != nil {
		log.Fatal("Failed to get latest block on startup:", err)
	}

	var startBlock uint64
	if checkpoint == 0 {
		// First run ever, we jump to real-time.
		// log.Printf("No checkpoint found, starting real-time mode from block %d", latest)
		startBlock = latest
	} else {
		// Resume from where we stopped last time.
		startBlock = checkpoint + 1
		// log.Printf("Starting from checkpoint at block %d", startBlock)
	}

	var nextExpectedBlock = startBlock
	var lastKnown uint64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		latest, err := s.client.BlockNumber(ctx)
		if err != nil {
			log.Printf("Failed to get latest block: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if lastKnown == 0 {
			// from my understanding, eth chain can sometimes be reorged : https://www.cube.exchange/what-is/chain-reorganization
			if startBlock > latest {
				startBlock = latest
			}
			lastKnown = latest
		}

		for current := startBlock; current <= latest; current++ {
			s.blocks <- current
			startBlock = current + 1
		}

		if latest > lastKnown {
			lastKnown = latest
		}

		drained := true
		for drained {
			select {
			case ackedBlock := <-s.ackChan:
				s.processedCount++
				if ackedBlock == nextExpectedBlock {
					nextExpectedBlock++
					// Save checkpoint every 5 confirmed blocks
					if nextExpectedBlock%5 == 0 {
						// log.Printf("Saving checkpoint at block %d", nextExpectedBlock-1)
						_ = s.state.SaveCheckpoint(nextExpectedBlock - 1)
					}
				}
			default:
				drained = false
			}
		}

		// log.Printf("waiting for a new block...")
		time.Sleep(s.config.PollInterval)
	}
}

func (s *Service) TestBlockChan() chan<- uint64 { return s.blocks }
