package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/xuri/excelize/v2"
	"log"
	"time"
)

type Result struct {
	Title   string  `json:"title"`
	Address string  `json:"address"`
	Star    *string `json:"star"`
}

func main() {
	url := "https://map.naver.com/v5/?c=14118278.7057239,4517436.9503642,16.19,0,0,0,dh"
	searchInputXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-input-box/div/div/div/input"
	searchValue := "맛집"

	searchFrameXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-layout/div[1]/combined-search-list/salt-search-list/nm-external-frame-bridge/nm-iframe/iframe"
	scrollDivXpath := "/html/body/div[3]/div/div[2]/div[1]"

	listXpath := "//li[contains(normalize-space(@class),'_1EKsQ _12tNp') and not(contains(@class, '_3in-q'))]"
	listClickSelector := "div._3hn9q > a > div > div > span.place_bluelink.OXiLu"

	entryFrameXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-layout/div[2]/entry-layout/entry-place-bridge/div/nm-external-frame-bridge/nm-iframe/iframe"
	titleSelector := "#_title > span._3XamX"
	addressSelector := "#app-root > div > div > div > div:nth-child(6) > div > div.place_section.no_margin._18vYz > div > ul > li._1M_Iz._1aj6- > div > a > span._2yqUQ"
	startSelector := "#app-root > div > div > div > div.place_section.GCwOh > div._3uUKd._2z4r0 > div._20Ivz > span._1Y6hi._1A8_M > em"

	startTime := time.Now()
	log.Printf("[startTime], %s", startTime)

	browser := rod.New().MustConnect().NoDefaultDevice()
	page := browser.MustPage(url).MustWindowFullscreen()

	err := page.WaitLoad()
	if err != nil {
		log.Fatal(err)
	}

	err = page.MustElementX(searchInputXpath).WaitVisible()
	if err != nil {
		log.Fatal(err)
	}

	page.MustElementX(searchInputXpath).MustInput(searchValue)
	page.KeyActions().Type(input.Enter).MustDo()

	err = page.MustElementX(searchFrameXpath).MustFrame().WaitLoad()
	if err != nil {
		log.Fatal(err)
	}

	searchFrame := page.MustElementX(searchFrameXpath).MustFrame()
	scrollDiv := searchFrame.MustElementX(scrollDivXpath)

	result := make([]Result, 0, 300)

	for {
		for {
			prevValue := scrollDiv.MustEval(`() => { this.scrollTop = this.scrollHeight; return this.scrollHeight; }`).Int()
			time.Sleep(1 * time.Second)
			nextValue := scrollDiv.MustEval(`() => this.scrollHeight`).Int()

			log.Printf("[scroll] prevValue: %d, nextValue: %d\n", prevValue, nextValue)

			if prevValue == nextValue {
				scrollDiv.MustEval(`() => { this.scrollTop = 0 }`)
				break
			}
		}

		li := searchFrame.MustElementsX(listXpath)
		for i := 0; i < len(li); i++ {
			li[i].MustElement(listClickSelector).MustClick()

			entryFrame := page.MustElementX(entryFrameXpath).MustFrame()
			err = entryFrame.WaitLoad()
			if err != nil {
				return
			}

			entryFrame.MustElement(titleSelector).MustWaitVisible()
			entryFrame.MustElement(addressSelector).MustWaitVisible()

			title, _ := entryFrame.MustElement(titleSelector).Text()

			address, _ := entryFrame.MustElement(addressSelector).Text()

			hasStar, starElement, _ := entryFrame.Has(startSelector)
			var star *string

			if hasStar {
				data, _ := starElement.Text()
				star = &data
			} else {
				star = nil
			}

			log.Printf("[title]: %s", title)
			log.Printf("[address]: %s", address)

			result = unique(result, Result{
				Title:   title,
				Address: address,
				Star:    star,
			})

			time.Sleep(1 * time.Second)
		}

		a := searchFrame.MustElementsX("/html/body/div[3]/div/div[2]/div[2]/a").Last()
		err = a.WaitEnabled()

		classValue := a.MustAttribute("class")
		if *classValue == "_2bgjK _34lTS" {
			break
		} else {
			a.MustEval("() => { this.click() }")
		}
	}

	if err != nil {
		return
	}

	f := excelize.NewFile()

	index := f.NewSheet("Sheet1")

	for index, value := range result {
		titleCell, _ := excelize.CoordinatesToCellName(2, index+2)
		addressCell, _ := excelize.CoordinatesToCellName(3, index+2)
		starCell, _ := excelize.CoordinatesToCellName(4, index+2)

		log.Printf("[titleCell] - %s", titleCell)
		log.Printf("[addressCell] - %s", addressCell)
		log.Printf("[Title] - %s", value.Title)
		log.Printf("[Address] - %s", value.Address)

		err := f.SetCellValue("Sheet1", titleCell, value.Title)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetCellValue("Sheet1", addressCell, value.Address)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetCellValue("Sheet1", starCell, *value.Star)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}
	}

	f.SetActiveSheet(index)

	if err := f.SaveAs("./test.xlsx"); err != nil {
		fmt.Println(err)
	}

	log.Printf("[latency], %s", time.Since(startTime))
}

func unique(result []Result, value Result) []Result {
	isExist := false
	for _, v := range result {
		if v.Title == value.Title && v.Address == value.Address {
			isExist = true
		}
	}
	if !isExist {
		result = append(result, value)
	}
	return result
}
