package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckpointState(t *testing.T) {
	tmpDir := t.TempDir()
	checkpointFile := filepath.Join(tmpDir, "checkpoint.txt")

	config := Config{
		File: checkpointFile,
	}

	state := NewFromConfig(config)

	initialBlockNum := state.LoadCheckpoint()
	if initialBlockNum != 0 {
		t.Errorf("Expected 0 from non-existent checkpoint file, got %d", initialBlockNum)
	}

	expectedBlockNum := uint64(12345)
	err := state.SaveCheckpoint(expectedBlockNum)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	fileInfo, err := os.Stat(checkpointFile)
	if err != nil {
		t.Fatalf("Failed to stat checkpoint file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	if fileInfo.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %v, got %v", expectedPerm, fileInfo.Mode().Perm())
	}

	loadedBlockNum := state.LoadCheckpoint()
	if loadedBlockNum != expectedBlockNum {
		t.Errorf("Expected to load block number %d, got %d", expectedBlockNum, loadedBlockNum)
	}

	newBlockNum := uint64(67890)
	err = state.SaveCheckpoint(newBlockNum)
	if err != nil {
		t.Fatalf("Failed to update checkpoint: %v", err)
	}

	updatedBlockNum := state.LoadCheckpoint()
	if updatedBlockNum != newBlockNum {
		t.Errorf("Expected updated block number %d, got %d", newBlockNum, updatedBlockNum)
	}

	err = os.WriteFile(checkpointFile, []byte("not-a-number"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted data: %v", err)
	}

	corruptedBlockNum := state.LoadCheckpoint()
	if corruptedBlockNum != 0 {
		t.Errorf("Expected 0 from corrupted checkpoint file, got %d", corruptedBlockNum)
	}
}
