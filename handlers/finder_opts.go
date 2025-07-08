package handlers

type finderOpts struct {
	concurrency int
}

type FinderOpt func(opts *finderOpts)

func WithConcurrency(concurrency int) FinderOpt {
	const defaultConcurrency = 5
	return func(opts *finderOpts) {
		if concurrency == 0 {
			concurrency = defaultConcurrency
		}
		opts.concurrency = concurrency
	}
}
