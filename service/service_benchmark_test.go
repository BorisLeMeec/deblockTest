package service

import (
	"context"
	"math/big"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"deblockTest/addressBook"
	"deblockTest/pkg"
	"deblockTest/service/mocks"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/mock/gomock"
)

var (
	testKey, _ = crypto.HexToECDSA("fad9c8855b548a0bb5e55f2197e1d0e5e9e8c4b9f5e4d3c2b1a0f9e8d7c6b5a4")
	testSigner = types.LatestSignerForChainID(big.NewInt(1))

	ab      *addressBook.AddressBook
	localAB = make([]common.Address, 500_000)
)

func init() {
	// Build real address book with 500,000 addresses
	cfg := &addressBook.Config{
		BloomExpected: 600_000,
		BloomFalsePos: 0.0001,
	}
	ab = addressBook.NewFromConfig(cfg)

	addresses := make(map[common.Address]string, 500_000)
	for i := 0; i < 500_000; i++ {
		addr := common.BytesToAddress(crypto.Keccak256([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})[12:])
		addresses[addr] = "user-" + string(rune(i))
		localAB[i] = addr
	}
	ab.SetAddresses(addresses)
}

func makeRealisticBlock(number uint64) *types.Block {
	txs := make([]*types.Transaction, 120)
	for i := 0; i < 120; i++ {
		var to common.Address
		if i%10 == 0 {
			to = localAB[rand.Intn(len(localAB))]
		} else {
			to = common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		}

		tx := types.NewTx(&types.LegacyTx{
			Nonce:    uint64(i),
			To:       &to,
			Value:    big.NewInt(1e15),
			Gas:      21000,
			GasPrice: big.NewInt(20e9),
		})
		signed, _ := types.SignTx(tx, testSigner, testKey)
		txs[i] = signed
	}
	header := &types.Header{Number: big.NewInt(int64(number))}
	return types.NewBlockWithHeader(header).WithBody(types.Body{Transactions: txs})
}

func BenchmarkService_RealThroughput(b *testing.B) {
	ctrl := gomock.NewController(b)

	ethMock := mocks.NewMockEthereumBlockGetter(ctrl)
	ethMock.EXPECT().BlockNumber(gomock.Any()).Return(uint64(20000000), nil).AnyTimes()
	ethMock.EXPECT().BlockByNumber(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n *big.Int) (*types.Block, error) {
			return makeRealisticBlock(n.Uint64()), nil
		},
	).AnyTimes()

	stateMock := mocks.NewMockState(ctrl)
	stateMock.EXPECT().LoadCheckpoint().Return(uint64(19999000)).Times(1)
	stateMock.EXPECT().SaveCheckpoint(gomock.Any()).Return(nil).AnyTimes()

	// Zero-cost mocks
	pub := &nopPublisher{}

	svc := NewService(
		&Config{
			PollInterval: 0,
			WorkerCount:  32,
		},
		ethMock, ab, pub, stateMock,
	)

	ctx, cancel := context.WithCancel(context.Background())
	svc.Setup(ctx)
	go svc.Run(ctx)

	time.Sleep(50 * time.Millisecond)

	b.ResetTimer()
	startTime := time.Now()

	for i := 0; i < b.N; i++ {
		blockNum := uint64(19999001 + i)
		svc.TestBlockChan() <- blockNum
	}

	for len(svc.TestBlockChan()) > 0 || atomic.LoadUint64(&svc.processedCount) < uint64(b.N) {
		time.Sleep(1 * time.Millisecond)
	}

	elapsed := time.Since(startTime)
	b.StopTimer()

	cancel()

	blocksPerSec := float64(b.N) / elapsed.Seconds()
	txPerSec := blocksPerSec * 120

	b.ReportMetric(blocksPerSec, "blocks/sec")
	b.ReportMetric(txPerSec, "tx/sec")
	b.ReportMetric(float64(b.N), "total_blocks")
	b.ReportMetric(elapsed.Seconds(), "total_seconds")
}

type nopPublisher struct{}

func (n *nopPublisher) Publish(context.Context, []pkg.TxMessage) {}
