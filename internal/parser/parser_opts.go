package parser

import "github.com/d1vbyz3r0/typed/internal/parser/request"

type ParseOpt func(p *parserOpts)

type parserOpts struct {
	parseAllModels         bool
	parseEnums             bool
	parseInlinePathParams  bool
	parseInlineQueryParams bool
	parseInlineForms       bool
}

func (o *parserOpts) RequestParseOpts() []request.ParseOpt {
	opts := make([]request.ParseOpt, 0, 3)
	if o.parseInlineQueryParams {
		opts = append(opts, request.ParseInlineQueryParams())
	}

	if o.parseInlineForms {
		opts = append(opts, request.ParseInlineForms())
	}

	if o.parseInlinePathParams {
		opts = append(opts, request.ParseInlinePathParams())
	}

	return opts
}

// ParseAllModels will allow you to parse all models in package used by echo.Bind call and declared in package
func ParseAllModels() ParseOpt {
	return func(p *parserOpts) {
		p.parseAllModels = true
	}
}

func ParseEnums() ParseOpt {
	return func(p *parserOpts) {
		p.parseEnums = true
	}
}

func ParseInlinePathParams() ParseOpt {
	return func(p *parserOpts) {
		p.parseInlinePathParams = true
	}
}

func ParseInlineQueryParams() ParseOpt {
	return func(p *parserOpts) {
		p.parseInlineQueryParams = true
	}
}

func ParseInlineForms() ParseOpt {
	return func(p *parserOpts) {
		p.parseInlineForms = true
	}
}
