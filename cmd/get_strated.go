package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/utils"
	"time"
)

func main() {
	browser := rod.New().MustConnect().NoDefaultDevice()
	page := browser.MustPage("https://en.wikipedia.org/").MustWindowFullscreen()

	page.MustElement("#searchInput").MustInput("earth")
	page.MustElement("#searchButton").MustClick()

	page.MustWaitLoad().MustScreenshot("a.png")
	el := page.MustElement("#mw-content-text > div.mw-parser-output > p:nth-child(9)")
	fmt.Println(el.MustText())

	el = page.MustElement("#mw-content-text > div.mw-parser-output > table.infobox > tbody > tr:nth-child(1) > td > a > img")
	_ = utils.OutputFile("b.png", el.MustResource())

	time.Sleep(time.Hour)
}
