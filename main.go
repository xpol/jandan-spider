package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const start = "http://jandan.net/ooxx"
const pageUrlTemplate = "http://jandan.net/ooxx/page-%d"

var waitGroup sync.WaitGroup

func textNumber(selection * goquery.Selection, selector string, page string, index int) (int) {
	html, err := selection.Html()
	text := selection.Find(selector).First().Text()
	v, err := strconv.Atoi(text)
	if err != nil {
		log.Panicf("%s(%d): %s\n%s", page, index, err, html)
	}
	return v
}

func Image(image* goquery.Selection) {
	defer waitGroup.Done()
	href, exists := image.Attr("href")
	if !exists {
		log.Panic("Image href not exists.")
	}

	log.Println(href)
}

func Post(post* goquery.Selection, page string, index int) {
	defer waitGroup.Done()
	if post.Length() == 0 {
		log.Print("Empty post")
		return
	}

	vote := post.Find(".jandan-vote").First()
	if vote.Length() == 0 {
		log.Print("Empty vote")
		return
	}

	like := textNumber(vote, "span.tucao-like-container > span", page, index)
	unlike := textNumber(vote, "span.tucao-unlike-container > span", page, index)

	if like < 100 || unlike * 8 > like {
		return
	}

	images := post.Find("a.view_img_link")
	waitGroup.Add(images.Length())
	log.Printf("Download %d images", images.Length())
	images.Each(func(_ int, image* goquery.Selection) {
		go Image(image)
	})
}

func Page(url string) {
	log.Printf("Fetching page: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}

	if resp.StatusCode != 200 {
		log.Panicf("Got status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Panic(err)
	}
	
	list := doc.Find("ol.commentlist").First()
	posts := list.Find("li")
	waitGroup.Add(posts.Length())
	posts.Each(func(i int, post* goquery.Selection) {
		go Post(post, url, i)
	})
}

func Start() {
	doc, err := goquery.NewDocument(start)
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile("[0-9]+")
	pageText := doc.Find("span.current-comment-page").First().Text()
	page, err := strconv.Atoi(re.FindString(pageText))
	if err != nil {
		log.Fatal(err)
	}

	for i:=1; i<=page; i++ {
		url := fmt.Sprintf(pageUrlTemplate, i)
		Page(url)
	}
	waitGroup.Wait()
}

func main() {
	Start()
}
