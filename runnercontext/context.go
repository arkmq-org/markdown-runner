package runnercontext

import (
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/view"
)

// Context provides a shared context for execution, containing
// the configuration and UI view.
type Context struct {
	Cfg *config.Config
	UI  view.RunnerView
}
