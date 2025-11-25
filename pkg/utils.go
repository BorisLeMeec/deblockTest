package pkg

import (
	"github.com/ethereum/go-ethereum/common"
)

func ToStringPtr(addr *common.Address) string {
	if addr == nil {
		return ""
	}
	return addr.Hex()
}
