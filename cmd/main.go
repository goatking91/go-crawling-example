package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/xuri/excelize/v2"
	"log"
	"strconv"
	"time"
	"unicode/utf8"
)

type Result struct {
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Address  string  `json:"address"`
	Star     *string `json:"star"`
}

func main() {
	url := "https://map.naver.com/v5/?c=14118278.7057239,4517436.9503642,16.19,0,0,0,dh"
	searchInputXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-input-box/div/div/div/input"
	searchValue := "식당"

	searchFrameXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-layout/div[1]/combined-search-list/salt-search-list/nm-external-frame-bridge/nm-iframe/iframe"
	scrollDivXpath := "/html/body/div[3]/div/div[2]/div[1]"

	listXpath := "//li[contains(normalize-space(@class),'_1EKsQ _12tNp') and not(contains(@class, '_3in-q'))]"
	listClickSelector := "div._3hn9q > a > div > div > span.place_bluelink.OXiLu"

	entryFrameXpath := "/html/body/app/layout/div[3]/div[2]/shrinkable-layout/div/app-base/search-layout/div[2]/entry-layout/entry-place-bridge/div/nm-external-frame-bridge/nm-iframe/iframe"
	titleSelector := "#_title > span._3XamX"
	categorySelector := "#_title > span._3ocDE"
	addressSelector := "#app-root > div > div > div > div:nth-child(6) > div > div.place_section.no_margin._18vYz > div > ul > li._1M_Iz._1aj6- > div > a > span._2yqUQ"
	starSelector := "#app-root > div > div > div > div.place_section.GCwOh > div._3uUKd._2z4r0 > div._20Ivz > span._1Y6hi._1A8_M > em"

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

			category, _ := entryFrame.MustElement(categorySelector).Text()

			address, _ := entryFrame.MustElement(addressSelector).Text()

			hasStar, starElement, _ := entryFrame.Has(starSelector)
			var star *string

			if hasStar {
				data, _ := starElement.Text()
				star = &data
			} else {
				data := ""
				star = &data
			}

			log.Printf("[title]: %s", title)
			log.Printf("[category]: %s", category)
			log.Printf("[address]: %s", address)
			log.Printf("[star]: %s", *star)

			result = unique(result, Result{
				Title:    title,
				Category: category,
				Address:  address,
				Star:     star,
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

	headerStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 14},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"DADEE0"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"}, {Type: "bottom", Style: 1, Color: "000000"},
			{Type: "left", Style: 1, Color: "000000"}, {Type: "right", Style: 1, Color: "000000"},
		},
	})

	bodyStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Size: 11},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"}, {Type: "bottom", Style: 1, Color: "000000"},
			{Type: "left", Style: 1, Color: "000000"}, {Type: "right", Style: 1, Color: "000000"},
		},
	})

	_ = f.SetCellValue("Sheet1", "B2", "순번")
	_ = f.SetCellValue("Sheet1", "C2", "상호명")
	_ = f.SetCellValue("Sheet1", "D2", "카테고리")
	_ = f.SetCellValue("Sheet1", "E2", "주소")
	_ = f.SetCellValue("Sheet1", "F2", "평점1")

	_ = f.SetCellStyle("Sheet1", "B2", "F2", headerStyle)

	for index, value := range result {
		noCell, _ := excelize.CoordinatesToCellName(2, index+3)
		titleCell, _ := excelize.CoordinatesToCellName(3, index+3)
		categoryCell, _ := excelize.CoordinatesToCellName(4, index+3)
		addressCell, _ := excelize.CoordinatesToCellName(5, index+3)
		starCell, _ := excelize.CoordinatesToCellName(6, index+3)

		err := f.SetCellValue("Sheet1", noCell, index+1)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetCellValue("Sheet1", titleCell, value.Title)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetCellValue("Sheet1", categoryCell, value.Category)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetCellValue("Sheet1", addressCell, value.Address)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		fl, err := strconv.ParseFloat(*value.Star, 64)
		if err != nil {
			err = f.SetCellValue("Sheet1", starCell, nil)
		} else {
			err = f.SetCellValue("Sheet1", starCell, fl)
		}

		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}
	}

	_ = f.SetCellStyle("Sheet1", "B3", "F102", bodyStyle)

	// Autofit all columns according to their text content
	cols, err := f.GetCols("Sheet1")
	if err != nil {
		log.Printf("[error] - %v", err)
		return
	}
	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell) + 16
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}

		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			log.Printf("[error] - %v", err)
			return
		}

		err = f.SetColWidth("Sheet1", name, name, float64(largestWidth))
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
