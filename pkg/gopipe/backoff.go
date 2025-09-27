package gopipe

import "github.com/mani-s-tiwari/gopipe/internal/utils"

// BackoffConfig re-export so users don't import internal/utils
type BackoffConfig = utils.BackoffConfig

// BackoffStrategy re-export
type BackoffStrategy = utils.BackoffStrategy

const (
	StrategyExponential = utils.StrategyExponential
	StrategyLinear      = utils.StrategyLinear
	StrategyConstant    = utils.StrategyConstant
	StrategyFibonacci   = utils.StrategyFibonacci
)

// DefaultBackoff returns a sensible default backoff configuration
func DefaultBackoff() BackoffConfig {
	return utils.DefaultBackoff()
}

// NewBackoff re-export
func NewBackoff(cfg BackoffConfig) *utils.Backoff {
	return utils.NewBackoff(cfg)
}
