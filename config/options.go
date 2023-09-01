package config

import "time"

type Options struct {
	placeholderReplace bool          // Enable placeholder replacement when rewriting the request path.
	pollLogInterval    time.Duration // polls matched log files every --poll_log_interval, or 250ms by default,  for newly created log pathnames.
}

func (opts *Options) SetPlaceholderReplace(flag bool) {
	opts.placeholderReplace = flag
}

func (opts *Options) EnablePlaceholderReplace() bool {
	return opts.placeholderReplace
}

func (opts *Options) SetPollLogInterval(interval time.Duration) {
	opts.pollLogInterval = interval
}

func (opts *Options) PollLogInterval() time.Duration {
	return opts.pollLogInterval
}
