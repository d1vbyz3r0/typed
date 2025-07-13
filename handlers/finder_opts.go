package handlers

type finderOpts struct {
	concurrency int
}

const defaultConcurrency = 5

func newFinderOpts() *finderOpts {
	return &finderOpts{
		concurrency: defaultConcurrency,
	}
}

type FinderOpt func(opts *finderOpts)

func WithConcurrency(concurrency int) FinderOpt {
	return func(opts *finderOpts) {
		if concurrency == 0 {
			concurrency = defaultConcurrency
		}
		opts.concurrency = concurrency
	}
}
