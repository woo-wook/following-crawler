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

	var f func(*html.Node)

	f = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "span" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == "f4 Link--primary" {
					if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
						fmt.Println(node.FirstChild.Data)
					}

					if node.Parent != nil && node.Parent.Data == "a" {
						for _, parentAttr := range node.Parent.Attr {
							if parentAttr.Key == "href" {
								user := parentAttr.Val
								urls = append(urls, "https://github.com"+user+"?tab=following")
								fmt.Println(urls)
								break
							}
						}
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return urls
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
