package stage

import (
	"os"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	pterm.DisableOutput()
	code := m.Run()
	pterm.EnableOutput()
	os.Exit(code)
}

func TestStage(t *testing.T) {
	t.Run("new stage", func(t *testing.T) {
		t.Run("creates a stage from chunks", func(t *testing.T) {
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage"},
				{Stage: "test-stage"},
			}
			cfg := &config.Config{}
			stage := NewStage(cfg, chunks)
			assert.NotNil(t, stage)
			assert.Equal(t, "test-stage", stage.Name)
			assert.Len(t, stage.Chunks, 2)
		})

		t.Run("returns nil for empty chunk slice", func(t *testing.T) {
			cfg := &config.Config{}
			stage := NewStage(cfg, []*chunk.ExecutableChunk{})
			assert.Nil(t, stage)
		})
	})
	t.Run("is parallelism consistent", func(t *testing.T) {
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
				assert.Equal(t, tt.expected, stage.IsParallelismConsistent())
			})
		}
	})
	t.Run("find chunk by id", func(t *testing.T) {
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
			assert.NotNil(t, foundChunk)
			assert.Equal(t, "chunk2", foundChunk.Id)
		})

		t.Run("returns nil for non-existent chunk id", func(t *testing.T) {
			foundChunk := FindChunkById(stages, "stage1", "non-existent")
			assert.Nil(t, foundChunk)
		})

		t.Run("returns nil for non-existent stage name", func(t *testing.T) {
			foundChunk := FindChunkById(stages, "non-existent", "chunk1")
			assert.Nil(t, foundChunk)
		})
	})
	t.Run("execute", func(t *testing.T) {
		t.Run("should execute a sequential stage", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"echo 1"}, Cfg: cfg},
				{Stage: "test-stage", Content: []string{"echo 2"}, Cfg: cfg},
			}
			stage := NewStage(cfg, chunks)
			err := stage.Execute(nil, make(map[string]string))
			assert.NoError(t, err)
		})

		t.Run("should execute a parallel stage", func(t *testing.T) {
			t.Setenv("CI", "true")
			cfg := &config.Config{MinutesToTimeout: 1}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Cfg: cfg},
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Cfg: cfg},
			}
			stage := NewStage(cfg, chunks)
			err := stage.Execute(nil, make(map[string]string))
			assert.NoError(t, err)
		})

		t.Run("should handle dependencies", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			stages := []*Stage{
				{
					Name: "setup",
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk1", Stage: "setup", Content: []string{"true"}, Cfg: cfg},
					},
				},
				{
					Name: "test-stage",
					Chunks: []*chunk.ExecutableChunk{
						{Stage: "test-stage", Requires: "setup/chunk1", Content: []string{"echo 'I have a dependency'"}, Cfg: cfg},
					},
				},
			}

			// First, execute the setup stage to simulate a real run
			setupStage := stages[0]
			setupStage.Cfg = cfg
			err := setupStage.Execute(stages, make(map[string]string))
			assert.NoError(t, err)

			// Then, execute the stage with the dependency
			testStage := stages[1]
			testStage.Cfg = cfg
			err = testStage.Execute(stages, make(map[string]string))
			assert.NoError(t, err)
		})

		t.Run("should handle breakpoints", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1, IgnoreBreakpoints: false}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", HasBreakpoint: true, Cfg: cfg},
			}
			stage := NewStage(cfg, chunks)
			err := stage.Execute(nil, make(map[string]string))
			assert.NoError(t, err)
			assert.True(t, cfg.Interactive)
		})
	})
	t.Run("execute with errors", func(t *testing.T) {
		t.Run("should stop execution on sequential stage error", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"false"}, Cfg: cfg},
				{Stage: "test-stage", Content: []string{"echo 'should not run'"}, Cfg: cfg},
			}
			stage := NewStage(cfg, chunks)
			err := stage.Execute(nil, make(map[string]string))
			assert.Error(t, err)
		})

		t.Run("should run all chunks on parallel stage error", func(t *testing.T) {
			t.Setenv("CI", "true")
			cfg := &config.Config{MinutesToTimeout: 1}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"false"}, IsParallel: true, Cfg: cfg},
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Cfg: cfg},
			}
			stage := NewStage(cfg, chunks)
			err := stage.Execute(nil, make(map[string]string))
			assert.Error(t, err)
		})

		t.Run("should skip chunk with unmet dependency", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			stages := []*Stage{
				{
					Name: "setup",
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk1", Stage: "setup", Content: []string{"false"}, Cfg: cfg}, // This chunk will fail
					},
				},
				{
					Name: "test-stage",
					Chunks: []*chunk.ExecutableChunk{
						{Stage: "test-stage", Requires: "setup/chunk1", Content: []string{"echo 'should not run'"}, Cfg: cfg},
					},
				},
			}

			setupStage := stages[0]
			setupStage.Cfg = cfg
			err := setupStage.Execute(stages, make(map[string]string))
			assert.Error(t, err)

			testStage := stages[1]
			testStage.Cfg = cfg
			err = testStage.Execute(stages, make(map[string]string))
			assert.NoError(t, err)
		})
	})
}
