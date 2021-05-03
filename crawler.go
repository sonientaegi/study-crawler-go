package main

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"study-crawler-go/utils"
	"sync"
	"time"
)

func fetch(url string) *html.Node {
	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		resp.Body.Close()
	}()

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

func crawl(userName string) (bool, []string) {
	var url = fmt.Sprintf("https://github.com/%s?tab=followers", userName)
	var node = fetch(url)
	if node != nil {
		return true, parseFollow(node)
	} else {
		return false, nil
	}
}

func worker(id string, chanRequest chan string, chanResponse chan<- string, chanTerminate <-chan struct{}) {
	for {
		select {
		case <-chanTerminate:
			return
		case requset := <-chanRequest:
			if ok, userNames := crawl(requset); ok && len(userNames) > 0 {
				for _, userName := range userNames {
					visited.m.Lock()
					if visited.userNames[userName] {
						visited.m.Unlock()
						continue
					}

					visited.userNames[userName] = true
					visited.m.Unlock()

					chanResponse <- userName

					select {
					case chanRequest <- userName:
					default:
						time.Sleep(time.Millisecond)
					}
				}
			} else {
				time.Sleep(time.Second * 1)
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

const NUM_OF_WORKERS = 8
const NUM_OF_MAX_RESULT = 10000
const MAX_QUEUE = NUM_OF_WORKERS * 50

var client = &http.Client{
	Timeout: time.Second * 5,
}

type _visited struct {
	m         sync.Mutex
	userNames map[string]bool
}

var visited = _visited{
	userNames: make(map[string]bool),
}

func main() {
	var wg sync.WaitGroup
	var chanRequest = make(chan string, MAX_QUEUE)
	var chanResponse = make(chan string)
	var chanTerminate = make(chan struct{})

	wg.Add(NUM_OF_WORKERS)
	chanRequest <- "sonientaegi"

	var ctx = context.Background()
	for i := 0; i < NUM_OF_WORKERS; i++ {
		var labels = pprof.Labels("ID", fmt.Sprintf("Worker - %03d", i))
		var f = func(ctx context.Context) {
			id, _ := pprof.Label(ctx, "ID")
			worker(id, chanRequest, chanResponse, chanTerminate)
			wg.Done()
		}
		go pprof.Do(ctx, labels, f)
	}

	totalProcessed := 0
	go func() {
		for _ = range chanResponse {
			totalProcessed++
			// log.Printf("%10d - %s", totalProcessed, resp)
		}
	}()

	time.Sleep(time.Second * 3)
	for {
		time.Sleep(time.Millisecond * 100)
		log.Printf("[%5d/%5d] Queue size = %d", totalProcessed, NUM_OF_MAX_RESULT, len(chanRequest))
		if totalProcessed >= NUM_OF_MAX_RESULT || len(chanRequest) == 0 {
			close(chanTerminate)
			break
		}
	}

	wg.Wait()
	close(chanResponse)
	close(chanRequest)
}
