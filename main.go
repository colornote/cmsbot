package main

import (
	"bytes"
	"cmsbot/database"
	"cmsbot/striptags"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"path"
	"path/filepath"
	"time"

	// "regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

type BlogArticle struct {
	Title    string `json:"title"`
	Links    string `json:"links"`
	Contents string `json:"contents"`
	// AuthorName string `gorm:"authorName"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.SetupDatabase()

	var p = flag.String("path", "./data", "The path to store html file.")
	flag.Parse()

	files, err := ioutil.ReadDir(*p)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() || path.Ext(file.Name()) != ".html" {
			continue
		}
		f, err := ioutil.ReadFile(filepath.Join(*p, file.Name()))

		if err != nil {
			panic(err)
		}

		Parser(f)

		fmt.Println(file.Name())
	}

	// Parser()
	fmt.Println("procesing done")
}

func Parser(file []byte) {

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(file))

	doc.Find(".js_album_item").Each(func(i int, s *goquery.Selection) {

		link, _ := s.Attr("data-link")
		title, _ := s.Attr("data-title")

		content := Fetch(link)

		if len(content) == 0 {
			return
		}

		var bl BlogArticle
		database.DB.Table("blog_article").Where("title = ?", title).Assign(BlogArticle{
			Title:    title,
			Links:    GenLink(),
			Contents: content,
			// AuthorName: "",
		}).FirstOrCreate(&bl)

		println(i, link, title)

	})

}

func GenLink() string {
	rand.Seed(time.Now().UnixNano())
	collect := []rune("abcedghijklnmopqrstuvwxyz-ABCEDGHIJKLNMOPQRSTUVWXYZ")
	ret := make([]rune, 1)
	for i := 15; i > 0; i-- {
		index := rand.Intn(len(collect))
		ret = append(ret, collect[index])
	}

	return fmt.Sprintf("%s.html", string(ret[:]))
}

func Fetch(url string) string {

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

	// fmt.Println("ss", resp.Body())

	if resp.StatusCode() == 200 {

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))

		// fmt.Println(htmlquery.OutputHTML(doc, false))

		div := doc.Find("#js_content")

		// fmt.Printf("%v", div.Text())

		if len(div.Text()) > 0 {
			strip_tags := striptags.NewStripTags()
			h, _ := div.Html()
			html_clean, err := strip_tags.Fetch(h) // 返回过滤后的 HTML 字符串和错误信息
			if err != nil {
				fmt.Printf("clean err: %v", err)
			}

			return html_clean
		}

	}
	return ""
}
