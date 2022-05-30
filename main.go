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
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

type BlogArticle struct {
	Title       string    `json:"title"`
	Links       string    `json:"links"`
	Contents    string    `json:"contents"`
	Description string    `json:"description"`
	Keywords    string    `json:"keywords"`
	AuthorName  string    `gorm:"column:authorName"`
	CreatedAt   time.Time `gorm:"column:createdAt"`
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

		list := strings.Split(file.Name(), ".")

		Parser(f, list[0])

		// fmt.Println(file.Name())
	}

	// Parser()
	fmt.Println("procesing done")
}

func Parser(file []byte, keyword string) {

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(file))

	doc.Find(".js_album_item").Each(func(i int, s *goquery.Selection) {

		link, _ := s.Attr("data-link")
		title, _ := s.Attr("data-title")

		var cn int64
		database.DB.Table("blog_article").Where("title = ?", title).Count(&cn)

		if cn > 0 {
			println(title, "已经在库~~~")
			return
		}

		content := Fetch(link)

		if len(content) == 0 {
			return
		}

		bl := BlogArticle{
			Title:       title,
			Links:       GenLink(),
			Contents:    content,
			Keywords:    keyword,
			Description: "",
			AuthorName:  "",
			CreatedAt:   time.Now(),
		}
		database.DB.Table("blog_article").Where("title = ?", title).Assign().Create(&bl)

		println(i, link, title, "处理完成")

	})

}

func GenLink() string {
	rand.Seed(time.Now().UnixNano())
	const collect = "abcedghijklnmopqrstuvwxyz-ABCEDGHIJKLNMOPQRSTUVWXYZ--"
	ret := make([]byte, 15)
	for i := range ret {
		index := rand.Intn(len(collect))
		ret[i] = collect[index]
	}

	return fmt.Sprintf("%s.html", ret)
}

func Fetch(url string) string {

	resp, err := resty.New().R().SetHeaders(map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"accept-language":           "zh-CN,zh;q=0.9",
		"cache-control":             "max-age=0",
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
			strip_tags.TrimSpace = true

			html_clean, err := strip_tags.Fetch(h) // 返回过滤后的 HTML 字符串和错误信息
			if err != nil {
				fmt.Printf("clean err: %v", err)
			}

			return html_clean
		}

	}
	return ""
}
