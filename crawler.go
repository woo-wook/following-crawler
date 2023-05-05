package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"runtime"
	"sync"
)

var fetched = struct {
	m map[string]error
	sync.Mutex
}{m: make(map[string]error)}

func fetch(url string) (*html.Node, error) {
	response, err := http.Get(url)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	doc, err := html.Parse(response.Body)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return doc, nil
}

func parseFollowing(doc *html.Node) []string {
	urls := make([]string, 0)

	urls = findFollowingInfo(doc, urls)

	return urls
}

// 여기서, 무언가 특정 정보를 찾아서 처리하면, 크롤링을 하는 의미가 생긴다.
func findFollowingInfo(node *html.Node, urls []string) []string {
	if node.Type == html.ElementNode && node.Data == "span" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "f4 Link--primary" {
				printTextNode(node)
				urls = getFollowingUrls(node, urls)
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		urls = findFollowingInfo(c, urls)
	}

	return urls
}

func getFollowingUrls(node *html.Node, urls []string) []string {
	if node.Parent != nil && node.Parent.Data == "a" {
		for _, attr := range node.Parent.Attr {
			if attr.Key == "href" {
				user := attr.Val
				urls = append(urls, "https://github.com"+user+"?tab=following")
				fmt.Println(urls)
				break
			}
		}
	}

	return urls
}

func printTextNode(node *html.Node) {
	if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
		fmt.Println(node.FirstChild.Data)
	}
}

func crawl(url string) {
	fetched.Lock()

	if _, ok := fetched.m[url]; ok {
		fetched.Unlock()
		return
	}

	fetched.Unlock()

	doc, err := fetch(url)

	fetched.Lock()
	fetched.m[url] = err
	fetched.Unlock()

	urls := parseFollowing(doc)
	done := make(chan bool)

	for _, u := range urls {
		go func(url string) {
			crawl(url)
			done <- true
		}(u)
	}

	for i := 0; i < len(urls); i++ {
		<-done
	}
}

func main() {
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)

	crawl("https://github.com/woo-wook?tab=following")
}
