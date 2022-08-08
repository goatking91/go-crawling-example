package main

import (
	"context"
	"github.com/go-rod/rod"
	"time"
)

func main() {
	page := rod.New().MustConnect().MustPage()
	ctx, cancel := context.WithCancel(context.Background())
	pageWithCancel := page.Context(ctx)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	pageWithCancel.MustNavigate("https://github.com")
	pageWithCancel.MustElement("body")

}
