## deblockTest – Real-time Ethereum Transaction Monitor  
Simple, fast, production-ready indexer that watches addresses and publishes user transactions.  
How to run (real mode – needs Alchemy)  
first argument = your Alchemy API key  
go run ./cmd/main.go YOUR_ALCHEMY_KEY  

The service will:  
Load the last checkpoint (or start from latest block)  
Continuously poll new blocks   
Extract transactions involving watched addresses  
Publish TxMessage events   
  
Performance (Apple M1 Pro – macOS)  
Realistic benchmark with:  
  
500 000 watched addresses (real Bloom filter + map)  
10% hit rate (very aggressive worst-case)  
Full transaction parsing + ACK logic  
No network calls (RPC mocked)  
  
go test -bench=BenchmarkService_RealThroughput -benchtime=10s -cpu=8  
BenchmarkService_RealThroughput-8   12188   966730 ns/op  
                                    1034 blocks/sec   
                                    124130 tx/sec  
                                    12188 total_blocks  
  
$ go test -bench=BenchmarkService_RealThroughput -benchtime=10s -cpu=12  
BenchmarkService_RealThroughput-12  13542   886629 ns/op  
                                    1128 blocks/sec  
                                    135344 tx/sec  
                                    13542 total_blocks  
  
Result → 124k – 135k transactions per second on a laptop  
