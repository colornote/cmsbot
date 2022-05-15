package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	// "regexp"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
)

func main() {

	timer := time.NewTimer(1 * time.Second)
	for {
		timer.Reset(120 * time.Second) // 这样来复用 timer 和修改执行时间
		select {
		case <-timer.C:
			fmt.Println("发起请求: ", time.Now().Format("2006-01-02 15:04"))
			if !Fetch("") {
				return
			}

		}
	}

	fmt.Println("Exit")
}

func Parser() {

	file, err := ioutil.ReadFile("data/#海外社媒营销.html")

	if err != nil {
		panic(err)
	}

	doc, _ := htmlquery.Parse(bytes.NewReader(file))

	// reg := regexp.MustCompile(`https.*300`)

	// list := htmlquery.Find(doc, "//div/@style")
	// for _, n := range list {

	// 	// println(n)
	// 	style := htmlquery.SelectAttr(n, "style")

	// 	//根据规则提取关键信息
	// 	result := reg.FindAllStringSubmatch(style, -1)
	// 	if len(result) == 0 {
	// 		continue
	// 	}
	// 	fmt.Println("result = ", result[0][0])

	// }

	list := htmlquery.Find(doc, "//li/@data-link")
	for _, n := range list {

		link := htmlquery.SelectAttr(n, "data-link")

		println(link)

	}
}

func Fetch(url string) bool {

	resp, err := resty.New().R().SetHeaders(map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"accept-language":           "zh-CN,zh;q=0.9",
		"cache-control":             "max-age=0",
		"if-modified-since":         "Sat, 14 May 2022 03:21:02 +0800",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "none",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"user-agent":                uarand.GetRandom(),
	}).Get(url)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode() == 200 {

		doc, _ := htmlquery.Parse(bytes.NewReader(resp.Body()))

		// fmt.Println(htmlquery.OutputHTML(doc, false))

		div := htmlquery.FindOne(doc, "//div[@id='js_content']")

		if div != nil {
			fmt.Printf("(%s)\n", htmlquery.OutputHTML(div, true))
		}

		link := htmlquery.SelectAttr(div, "url")

		println(link)

	} else {
		fmt.Println("访问失败: " + resp.Status())
	}
	return true
}
