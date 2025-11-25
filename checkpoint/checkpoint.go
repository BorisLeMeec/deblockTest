package checkpoint

import (
	"fmt"
	"os"
)

type State struct {
	config Config
}

func NewFromConfig(config Config) *State {
	return &State{config: config}
}

func (s *State) LoadCheckpoint() uint64 {
	data, err := os.ReadFile(s.config.File)
	if err != nil {
		return 0
	}
	var n uint64
	_, err = fmt.Sscanf(string(data), "%d", &n)
	if err != nil {
		return 0
	}
	return n
}

func (s *State) SaveCheckpoint(blockNumber uint64) error {
	return os.WriteFile(s.config.File, []byte(fmt.Sprintf("%d\n", blockNumber)), 0644)
}
