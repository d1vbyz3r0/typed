package request

type ParseOpt func(opts *requestParseOpts)

type requestParseOpts struct {
	parseInlinePathParams  bool
	parseInlineQueryParams bool
	parseInlineForms       bool
	parseInlineHeaders     bool
}

func ParseInlinePathParams() ParseOpt {
	return func(opts *requestParseOpts) {
		opts.parseInlinePathParams = true
	}
}

func ParseInlineQueryParams() ParseOpt {
	return func(opts *requestParseOpts) {
		opts.parseInlineQueryParams = true
	}
}

func ParseInlineForms() ParseOpt {
	return func(opts *requestParseOpts) {
		opts.parseInlineForms = true
	}
}

func ParseInlineHeaders() ParseOpt {
	return func(opts *requestParseOpts) {
		opts.parseInlineHeaders = true
	}
}
