// main.go
package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"deblockTest/addressBook"
	"deblockTest/checkpoint"
	"deblockTest/kafka"
	service2 "deblockTest/service"
)

const (
	rpcURL         = "https://eth-mainnet.g.alchemy.com/v2/"
	kafkaBroker    = "localhost:9092"
	kafkaTopic     = "eth-transactions"
	checkpointFile = "checkpoint.txt"
	workerCount    = 4
	bloomExpected  = 600_000 // slighty larger than number of addresses, to keep some bit at 0 (otherwise 100% false positive).
	bloomFalsePos  = 0.0001
	pollInterval   = 1 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		cancel()
	}()

	// Setup
	addresses := loadAddresses()
	ab := addressBook.NewFromConfig(&addressBook.Config{
		BloomExpected: bloomExpected,
		BloomFalsePos: bloomFalsePos,
	})
	ab.SetAddresses(addresses)

	k := kafka.NewFromConfig(&kafka.Config{Broker: kafkaBroker, Topic: kafkaTopic})
	defer k.Close()

	if len(os.Args) < 2 {
		log.Fatal("missing eth api key")
	}

	client, err := ethclient.Dial(rpcURL + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	s := checkpoint.NewFromConfig(checkpoint.Config{File: checkpointFile})

	service := service2.NewService(
		&service2.Config{PollInterval: pollInterval, WorkerCount: workerCount, CheckpointFile: checkpointFile},
		client,
		ab,
		k,
		s,
	)

	service.Setup(ctx)
	service.Run(ctx)
}

func loadAddresses() map[common.Address]string {
	// Simulate 500k addresses
	m := make(map[common.Address]string, 500_000)
	for i := 0; i < 500_000; i++ {
		addr := common.HexToAddress(fmt.Sprintf("0x%040x", i+1))
		m[addr] = fmt.Sprintf("user-%d", i+1)
	}
	log.Printf("Loaded %d addresses", len(m))

	return m
}
