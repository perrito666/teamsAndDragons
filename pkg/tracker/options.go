// Package tracker provides the public API for the career progression tracker.
package tracker

import (
	"log/slog"
)

// Options configures the Tracker.
type Options struct {
	// DataDir is the directory where person data is stored.
	// Defaults to "./data".
	DataDir string

	// Logger is the structured logger to use.
	// Defaults to slog.Default().
	Logger *slog.Logger
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		DataDir: "./data",
		Logger:  slog.Default(),
	}
}

// Option is a functional option for configuring a Tracker.
type Option func(*Options)

// WithDataDir sets the data directory.
func WithDataDir(dir string) Option {
	return func(o *Options) {
		o.DataDir = dir
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}
