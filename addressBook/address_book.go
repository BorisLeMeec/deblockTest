package addressBook

import (
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/ethereum/go-ethereum/common"
)

type AddressBook struct {
	bloom         *bloom.BloomFilter
	addressToUser map[common.Address]string
	config        *Config
}

func NewFromConfig(cfg *Config) *AddressBook {
	return &AddressBook{
		config: cfg,
		bloom:  bloom.NewWithEstimates(cfg.BloomExpected, cfg.BloomFalsePos),
	}
}

func (ab *AddressBook) SetAddresses(addresses map[common.Address]string) {
	ab.addressToUser = make(map[common.Address]string, len(addresses))

	for addr, userID := range addresses {
		ab.bloom.Add(addr.Bytes())
		ab.addressToUser[addr] = userID
	}
}

func (ab *AddressBook) GetUserID(addr common.Address) (string, bool) {
	if !ab.bloom.Test(addr.Bytes()) {
		return "", false
	}
	userID, ok := ab.addressToUser[addr]
	return userID, ok
}
