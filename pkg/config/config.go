package config

import (
	"flag"
	"time"
)

// Seed is the seed to use for RNGs.
var Seed int64

// InitConfigFlags sets up the config CLI flags.
func InitConfigFlags() {
	flag.Int64Var(&Seed, "seed", time.Now().UnixNano(), "The seed for the RNG.")
}
