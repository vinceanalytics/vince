// Copyright 2022 Molecula Corp. (DBA FeatureBase).
// SPDX-License-Identifier: Apache-2.0
package cfg

import "log/slog"

const (
	DefaultMinWALCheckpointSize = 1 * (1 << 20) // 1MB
	DefaultMaxWALCheckpointSize = DefaultMaxWALSize / 2
)

// Config defines externally configurable rbf options.
// The separate package avoids circular import.
type Config struct {

	// The maximum allowed database size. Required by mmap.
	MaxSize int64 `toml:"max-db-size"`

	// The maximum allowed WAL size. Required by mmap.
	MaxWALSize int64 `toml:"max-wal-size"`

	// The minimum WAL size before the WAL is copied to the DB.
	MinWALCheckpointSize int64 `toml:"min-wal-checkpoint-size"`

	// The maximum WAL size before transactions are halted to allow a checkpoint.
	MaxWALCheckpointSize int64 `toml:"max-wal-checkpoint-size"`

	// Set before calling db.Open()
	FsyncEnabled    bool `toml:"fsync"`
	FsyncWALEnabled bool `toml:"fsync-wal"`

	// for mmap correctness testing.
	DoAllocZero bool `toml:"do-alloc-zero"`

	// CursorCacheSize is the number of copies of Cursor{} to keep in our
	// readyCursorCh arena to avoid GC pressure.
	CursorCacheSize int64 `toml:"cursor-cache-size"`

	// Logger specifies a logger for asynchronous errors, such as
	// background checkpoints. It cannot be set from toml. The default is
	// to use stderr.
	Logger *slog.Logger `toml:"-"`

	// The maximum number of bits to be deleted in a single transaction default(65536)
	MaxDelete int `toml:"max-delete"`
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxSize:              DefaultMaxSize,
		MaxWALSize:           DefaultMaxWALSize,
		MinWALCheckpointSize: DefaultMinWALCheckpointSize,
		MaxWALCheckpointSize: DefaultMaxWALCheckpointSize,
		FsyncEnabled:         true,
		FsyncWALEnabled:      true,
		MaxDelete:            DefaultMaxDelete,

		// CI passed with 20. 50 was too big for CI, even on X-large instances.
		// For now we default to 0, which means use sync.Pool.
		CursorCacheSize: 0,
	}
}
