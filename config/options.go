package config

type Options struct {
	placeholderReplace bool // Enable placeholder replacement when rewriting the request path.
}

func (opts *Options) SetPlaceholderReplace(flag bool) {
	opts.placeholderReplace = flag
}

func (opts *Options) EnablePlaceholderReplace() bool {
	return opts.placeholderReplace
}
