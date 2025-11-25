package service

import "time"

type Config struct {
	WorkerCount    int
	PollInterval   time.Duration
	CheckpointFile string
}
