package addressBook

import (
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestAddressBook(t *testing.T) {
	cfg := &Config{
		BloomExpected: 1000,
		BloomFalsePos: 0.01,
	}

	ab := NewFromConfig(cfg)
	if ab == nil {
		t.Fatal("Failed to create address book")
	}

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	userID, found := ab.GetUserID(addr)
	if found {
		t.Errorf("Expected address not to be found in empty book, but got userID: %s", userID)
	}

	addresses := map[common.Address]string{
		common.HexToAddress("0x1234567890123456789012345678901234567890"): "user1",
		common.HexToAddress("0x2345678901234567890123456789012345678901"): "user2",
		common.HexToAddress("0x3456789012345678901234567890123456789012"): "user3",
	}
	ab.SetAddresses(addresses)

	testCases := []struct {
		address    string
		expectedID string
		shouldFind bool
	}{
		{"0x1234567890123456789012345678901234567890", "user1", true},
		{"0x2345678901234567890123456789012345678901", "user2", true},
		{"0x3456789012345678901234567890123456789012", "user3", true},
		{"0x9999999999999999999999999999999999999999", "", false},
	}

	for _, tc := range testCases {
		addr := common.HexToAddress(tc.address)
		userID, found := ab.GetUserID(addr)

		if found != tc.shouldFind {
			t.Errorf("For address %s: expected found=%v, got found=%v", tc.address, tc.shouldFind, found)
		}

		if found && userID != tc.expectedID {
			t.Errorf("For address %s: expected userID=%s, got userID=%s", tc.address, tc.expectedID, userID)
		}
	}
}
