package rvnodegen

import "time"

type optionConfig struct {
	discoveryCacheDir string
	httpCacheDir      string
	discoveryTTL      time.Duration
}

func buildOptionConfig(options ...Option) optionConfig {
	opts := optionConfig{
		discoveryCacheDir: "",
		httpCacheDir:      "",
		discoveryTTL:      180 * time.Second,
	}

	for _, o := range options {
		o(&opts)
	}

	return opts
}

// Option are node gen options.
type Option func(o *optionConfig)

// DiscoveryCacheDir sets the discovery cache directory.
func DiscoveryCacheDir(dir string) Option {
	return func(o *optionConfig) {
		o.discoveryCacheDir = dir
	}
}

// HTTPCacheDir sets the HTTP cache directory.
func HTTPCacheDir(dir string) Option {
	return func(o *optionConfig) {
		o.httpCacheDir = dir
	}
}

// DiscoveryTTL sets the ttl for discovery.
func DiscoveryTTL(ttl time.Duration) Option {
	return func(o *optionConfig) {
		o.discoveryTTL = ttl
	}
}
