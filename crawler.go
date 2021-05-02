package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"study-crawler-go/utils"
	"sync"
	"time"
)

func fetch(url string) *html.Node {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	node, err := html.Parse(resp.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	return node
}

func fetchLocal(dir string) *html.Node {
	file, err := os.Open(dir)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer file.Close()

	node, err := html.Parse(file)
	if err != nil {
		log.Println(err)
		return nil
	}

	return node
}

type elementType struct {
	node          *html.Node
	isChildOfLink bool
	url           string
}

func parseFollow(node *html.Node) []string {
	var userNames []string = make([]string, 0)
	var queue = new(utils.Queue).Init()

	queue.Push(elementType{
		node:          node,
		isChildOfLink: false,
		url:           ""},
	)

TRAVERSAL:
	for {
		if queue.Len() == 0 {
			break TRAVERSAL
		}

		//if element.node.Type != html.ElementNode {
		//	continue TRAVERSAL
		//}

		var element = queue.Pop().(elementType)
		var url = ""
		var isChildOfLink = false

		// fmt.Println(element.node.Type, element.node.Data)

		switch element.node.Data {
		case "a":
			var dataHovercardType = ""
			var href = ""
			for _, attr := range element.node.Attr {
				switch attr.Key {
				case "data-hovercard-type":
					dataHovercardType = attr.Val
				case "href":
					href = attr.Val
				}
			}

			if dataHovercardType == "user" && href != "" {
				isChildOfLink = true
				url = href
			}
		case "img":
			if !element.isChildOfLink {
				break
			}

			for _, attr := range element.node.Attr {
				switch attr.Key {
				case "class":
					if attr.Val == "avatar avatar-user" {
						userNames = append(userNames, element.url[1:])
					}
				}
			}

		}

		for child := element.node.FirstChild; child != nil; child = child.NextSibling {
			queue.Push(elementType{
				node:          child,
				isChildOfLink: isChildOfLink,
				url:           url,
			})
		}
	}

	return userNames
}

const INITIAL_URL = "https://github.com/SonienTaegi?tab=followers"

var m sync.Mutex
var visited = make(map[string]bool)

func crawl(userName string) {
	m.Lock()
	if visited[userName] {
		m.Unlock()
		return
	}

	visited[userName] = true
	m.Unlock()

	var url = fmt.Sprintf("https://github.com/%s?tab=followers", userName)
	var done = make(chan bool)
	var followers = parseFollow(fetch(url))

	var msg = fmt.Sprintf("%20s : %d ", userName, len(followers))
	if len(followers) == 0 {
		file, _ := os.Create(userName + ".html")
		resp, err := http.Get(fmt.Sprintf("https://github.com/%s?tab=followers", userName))
		if err != nil {
			msg += fmt.Sprint(err)
			errBytes, _ := ioutil.ReadAll(strings.NewReader(fmt.Sprint(err)))
			file.Write(errBytes)
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			file.Write(body)
		}
		file.Close()
	}
	log.Print(msg)

	for _, follower := range followers {
		time.Sleep(time.Second)
		go func(follower string) {
			crawl(follower)
			done <- true
		}(follower)
	}

	for i := 0; i < len(followers); i++ {
		<-done
	}
}

func main() {
	crawl("sonientaegi")
}
