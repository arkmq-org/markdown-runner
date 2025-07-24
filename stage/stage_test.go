package stage

import (
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
)

func TestNewStage(t *testing.T) {
	t.Run("creates a stage from chunks", func(t *testing.T) {
		chunks := []*chunk.ExecutableChunk{
			{Stage: "test-stage"},
			{Stage: "test-stage"},
		}
		stage := NewStage(chunks)
		if stage == nil {
			t.Fatal("NewStage returned nil")
		}
		if stage.Name != "test-stage" {
			t.Errorf("expected stage name 'test-stage', got '%s'", stage.Name)
		}
		if len(stage.Chunks) != 2 {
			t.Errorf("expected 2 chunks, got %d", len(stage.Chunks))
		}
	})

	t.Run("returns nil for empty chunk slice", func(t *testing.T) {
		stage := NewStage([]*chunk.ExecutableChunk{})
		if stage != nil {
			t.Error("expected nil for empty chunk slice, got a stage")
		}
	})
}

func TestIsParallelismConsistent(t *testing.T) {
	tests := []struct {
		name     string
		chunks   []*chunk.ExecutableChunk
		expected bool
	}{
		{
			name: "all parallel",
			chunks: []*chunk.ExecutableChunk{
				{IsParallel: true},
				{IsParallel: true},
			},
			expected: true,
		},
		{
			name: "all sequential",
			chunks: []*chunk.ExecutableChunk{
				{IsParallel: false},
				{IsParallel: false},
			},
			expected: true,
		},
		{
			name: "mixed parallel and sequential",
			chunks: []*chunk.ExecutableChunk{
				{IsParallel: true},
				{IsParallel: false},
			},
			expected: false,
		},
		{
			name:     "no chunks",
			chunks:   []*chunk.ExecutableChunk{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := Stage{Chunks: tt.chunks}
			if got := stage.IsParallelismConsistent(); got != tt.expected {
				t.Errorf("isParallelismConsistent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindChunkById(t *testing.T) {
	stages := []*Stage{
		{
			Name: "stage1",
			Chunks: []*chunk.ExecutableChunk{
				{Id: "chunk1"},
				{Id: "chunk2"},
			},
		},
		{
			Name: "stage2",
			Chunks: []*chunk.ExecutableChunk{
				{Id: "chunk3"},
			},
		},
	}

	t.Run("finds existing chunk", func(t *testing.T) {
		foundChunk := FindChunkById(stages, "stage1", "chunk2")
		if foundChunk == nil {
			t.Fatal("expected to find chunk, but got nil")
		}
		if foundChunk.Id != "chunk2" {
			t.Errorf("expected chunk with id 'chunk2', got '%s'", foundChunk.Id)
		}
	})

	t.Run("returns nil for non-existent chunk id", func(t *testing.T) {
		foundChunk := FindChunkById(stages, "stage1", "non-existent")
		if foundChunk != nil {
			t.Error("expected not to find chunk, but got one")
		}
	})

	t.Run("returns nil for non-existent stage name", func(t *testing.T) {
		foundChunk := FindChunkById(stages, "non-existent", "chunk1")
		if foundChunk != nil {
			t.Error("expected not to find chunk, but got one")
		}
	})
}
