package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSleepLogNextToTheScores(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteSleepLog(filepath.Join(dir, ScoresFile))
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, SleepLogFile) {
		t.Errorf("written to %q, want it next to the scores file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "Stay that way. For me.") || !strings.Contains(string(data), "awake 3:17") {
		t.Errorf("file = %q, %v", data, err)
	}
	// Writing twice is fine (the same file is rewritten).
	if _, err := WriteSleepLog(filepath.Join(dir, ScoresFile)); err != nil {
		t.Error(err)
	}
}

func TestWriteSleepLogNeedsAScoresPath(t *testing.T) {
	if path, err := WriteSleepLog(""); path != "" || err != nil {
		t.Errorf("test mode must write nothing: %q, %v", path, err)
	}
}
