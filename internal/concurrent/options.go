package concurrent

type (
	concurrency struct {
		maxConcurrency int64
	}

	ConcurrencyOptions func(*concurrency)
)

const (
	defaultMaxConcurrency = 10
)

var (
	//nolint:gochecknoglobals // defaultOpts is safe to use as a package-level variable.
	defaultOpts = &concurrency{
		maxConcurrency: defaultMaxConcurrency,
	}
)

// WithMaxConcurrency sets the maximum concurrency for running tasks.
func WithMaxConcurrency(n int64) ConcurrencyOptions {
	return func(r *concurrency) {
		r.maxConcurrency = n
	}
}

func setOpts(options ...ConcurrencyOptions) *concurrency {
	opts := defaultOpts
	for _, o := range options {
		o(opts)
	}
	return opts
}
