package flagparser

import (
	"flag"
	"time"
)

const (
	defaultRecursiveCrawlDepth = 3
	defaultReqTimeoutSeconds   = 5
)

type fp struct {
	url               *string
	maxDepth          *uint
	reqTimeOutSeconds *uint
}

type FlagParser interface {
	// GetURL returns provided URL to parse
	GetURL() string
	// GetMaxDepth returns provided recursive crawl depth
	GetMaxDepth() uint
	// GetReqTimeout returns provided request timeout for an individual request
	GetReqTimeout() time.Duration
}

func NewFlagParser() FlagParser {
	f := fp{}

	f.url = flag.String("url", "",
		"provide URL to crawl")
	f.maxDepth = flag.Uint("max_depth", defaultRecursiveCrawlDepth,
		"max depth of recursive URL crawl, default is 10")
	f.reqTimeOutSeconds = flag.Uint("req_timeout_sec", defaultReqTimeoutSeconds,
		"timeout for an individual crawl request")

	flag.Parse()

	return &f
}

func (f *fp) GetURL() string {
	return *f.url
}

func (f *fp) GetMaxDepth() uint {
	return *f.maxDepth
}

func (f *fp) GetReqTimeout() time.Duration {
	return time.Duration(*f.reqTimeOutSeconds) * time.Second
}
