package main

import (
	"context"

	"github.com/tinbtb/rcrawl/internal/app"
)

func main() {
	app.NewApp().Run(context.Background())
}
