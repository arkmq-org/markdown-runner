package stage

import (
	"os"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/view"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestStage(t *testing.T) {
	t.Run("new stage", func(t *testing.T) {
		t.Run("creates a stage from chunks", func(t *testing.T) {
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage"},
				{Stage: "test-stage"},
			}
			cfg := &config.Config{}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			stage := NewStage(ctx, chunks)
			assert.NotNil(t, stage)
			assert.Equal(t, "test-stage", stage.Name)
			assert.Len(t, stage.Chunks, 2)
		})

		t.Run("returns nil for empty chunk slice", func(t *testing.T) {
			cfg := &config.Config{}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			stage := NewStage(ctx, []*chunk.ExecutableChunk{})
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
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"echo 1"}, Context: ctx},
				{Stage: "test-stage", Content: []string{"echo 2"}, Context: ctx},
			}
			stage := NewStage(ctx, chunks)
			err := stage.Execute(nil, make(map[string]string), nil)
			assert.NoError(t, err)
		})

		t.Run("should execute a parallel stage", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Context: ctx},
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Context: ctx},
			}
			stage := NewStage(ctx, chunks)
			err := stage.Execute(nil, make(map[string]string), nil)
			assert.NoError(t, err)
		})

		t.Run("should handle dependencies", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			stages := []*Stage{
				{
					Name: "setup",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk1", Stage: "setup", Content: []string{"true"}, Context: ctx},
					},
				},
				{
					Name: "test-stage",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Stage: "test-stage", Requires: "setup/chunk1", Content: []string{"echo 'I have a dependency'"}, Context: ctx},
					},
				},
			}

			// First, execute the setup stage to simulate a real run
			setupStage := stages[0]
			setupStage.Ctx = ctx
			err := setupStage.Execute(stages, make(map[string]string), nil)
			assert.NoError(t, err)

			// Then, execute the stage with the dependency
			testStage := stages[1]
			testStage.Ctx = ctx
			err = testStage.Execute(stages, make(map[string]string), nil)
			assert.NoError(t, err)
		})

		t.Run("should handle breakpoints", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1, IgnoreBreakpoints: false}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", HasBreakpoint: true, Context: ctx},
			}
			stage := NewStage(ctx, chunks)
			err := stage.Execute(nil, make(map[string]string), nil)
			assert.NoError(t, err)
			assert.True(t, cfg.Interactive)
		})
	})
	t.Run("execute with errors", func(t *testing.T) {
		t.Run("should not execute subsequent stages on failure unless it is a teardown stage", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			stages := []*Stage{
				{
					Name: "stage1",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk1_success", Stage: "stage1", Content: []string{"true"}, Context: ctx},
						{Id: "chunk1_fail", Stage: "stage1", Content: []string{"false"}, Context: ctx},
					},
				},
				{
					Name: "stage2",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk2", Stage: "stage2", Content: []string{"echo 'should not run'"}, Context: ctx},
					},
				},
				{
					Name: "teardown",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Id: "teardown_for_success", Requires: "stage1/chunk1_success", Stage: "teardown", Content: []string{"echo 'should run'"}, Context: ctx},
						{Id: "teardown_for_fail", Requires: "stage1/chunk1_fail", Stage: "teardown", Content: []string{"echo 'should not run'"}, Context: ctx},
					},
				},
			}

			var err error
			for _, s := range stages {
				err2 := s.Execute(stages, make(map[string]string), err)
				if err2 != nil {
					err = err2
				}
			}

			assert.Error(t, err)
			assert.True(t, stages[0].Chunks[0].HasExecutedCorrectly())
			assert.False(t, stages[0].Chunks[1].HasExecutedCorrectly())
			assert.True(t, stages[1].Chunks[0].IsSkipped)
			assert.True(t, stages[2].Chunks[0].HasExecutedCorrectly())
			assert.False(t, stages[2].Chunks[1].HasFinishedExecution())
		})

		t.Run("should stop execution on sequential stage error", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"false"}, Context: ctx},
				{Stage: "test-stage", Content: []string{"echo 'should not run'"}, Context: ctx},
			}
			stage := NewStage(ctx, chunks)
			err := stage.Execute(nil, make(map[string]string), nil)
			assert.Error(t, err)
		})

		t.Run("should run all chunks on parallel stage error", func(t *testing.T) {
			t.Setenv("CI", "true")
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			chunks := []*chunk.ExecutableChunk{
				{Stage: "test-stage", Content: []string{"false"}, IsParallel: true, Context: ctx},
				{Stage: "test-stage", Content: []string{"sleep 0.1"}, IsParallel: true, Context: ctx},
			}
			stage := NewStage(ctx, chunks)
			err := stage.Execute(nil, make(map[string]string), nil)
			assert.Error(t, err)
		})

		t.Run("should skip chunk with unmet dependency", func(t *testing.T) {
			cfg := &config.Config{MinutesToTimeout: 1}
			ui := view.NewMock()
			ctx := &runnercontext.Context{
				Cfg: cfg,
				UI:  ui,
			}
			stages := []*Stage{
				{
					Name: "setup",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Id: "chunk1", Stage: "setup", Content: []string{"false"}, Context: ctx}, // This chunk will fail
					},
				},
				{
					Name: "test-stage",
					Ctx:  ctx,
					Chunks: []*chunk.ExecutableChunk{
						{Stage: "test-stage", Requires: "setup/chunk1", Content: []string{"echo 'should not run'"}, Context: ctx},
					},
				},
			}

			setupStage := stages[0]
			setupStage.Ctx = ctx
			err := setupStage.Execute(stages, make(map[string]string), nil)
			assert.Error(t, err)

			testStage := stages[1]
			testStage.Ctx = ctx
			err = testStage.Execute(stages, make(map[string]string), nil)
			assert.NoError(t, err)
		})
	})
}
