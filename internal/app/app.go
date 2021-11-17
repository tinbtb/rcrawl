package app

import (
	"context"
	"fmt"

	"github.com/tinbtb/rcrawl/internal/app/flagparser"
	"github.com/tinbtb/rcrawl/internal/app/urlcrawler"
	"github.com/tinbtb/rcrawl/pkg/signalhandler"
)

type app struct{}

type App interface {
	Run(context.Context)
}

func NewApp() App {
	return &app{}
}

func (app) Run(ctx context.Context) {
	ctx = signalhandler.NewSignalHandler().CatchSignals(ctx)

	fp := flagparser.NewFlagParser()

	uc := urlcrawler.NewURLCrawler(fp.GetReqTimeout())

	err := uc.Crawl(ctx, fp.GetURL(), fp.GetMaxDepth())
	if err != nil {
		// use your favorite logger here
		fmt.Println("Failed to crawl: " + err.Error())
	}

	fmt.Println("Finished.")
}
