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
	visited.m.Lock()
	if visited.userNames[userName] {
		visited.m.Unlock()
		return false, nil
	}

	visited.userNames[userName] = true
	visited.m.Unlock()

	var url = fmt.Sprintf("https://github.com/%s?tab=followers", userName)
	var node = fetch(url)
	if node != nil {
		return true, parseFollow(node)
	} else {
		return false, nil
	}
}

func worker(id string, chanRequest <-chan string, chanResponse chan<- string, chanTerminate <-chan bool) {
	for {
		// time.Sleep(time.Millisecond * 500)
		select {
		case <-chanTerminate:
			return
		case userName := <-chanRequest:
			if ok, userNames := crawl(userName); ok && len(userNames) > 0 {
				for _, u := range userNames {
					chanResponse <- u
				}
				time.Sleep(time.Second * 1)
			} else {
				log.Println(id, "wait ...")
				time.Sleep(time.Second * 10)
			}
		}
	}
}

const NUM_OF_WORKERS = 8
const NUM_OF_MAX_RESULT = 100000

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

type _requestQueue struct {
	m sync.Mutex
	q *utils.Queue
}

var requestQueue = _requestQueue{
	q: new(utils.Queue).Init(),
}

func main() {
	var wg sync.WaitGroup
	var chanRequest = make(chan string)
	var chanResponse = make(chan string)
	var chanTerminate = make(chan bool)

	wg.Add(NUM_OF_WORKERS)
	requestQueue.q.Push("sonientaegi")

	go func() {
		defer func() {
			recover()
		}()

		for {
			var userName string
			requestQueue.m.Lock()
			if requestQueue.q.Len() > 0 {
				userName = requestQueue.q.Pop().(string)
			}
			requestQueue.m.Unlock()

			if userName == "" {
				time.Sleep(time.Second)
				continue
			}

			select {
			case chanRequest <- userName:
				break
			}
		}
	}()

	go func() {
		var requestTermination = false
		var cnt = 0
		for userName := range chanResponse {
			if cnt == NUM_OF_MAX_RESULT {
				if !requestTermination {
					go func() {
						for i := 0; i < NUM_OF_WORKERS; i++ {
							chanTerminate <- true
						}
					}()
					requestTermination = true
				}
				continue
			}
			cnt++
			log.Printf("%10d : %s", cnt, userName)

			requestQueue.m.Lock()
			requestQueue.q.Push(userName)
			requestQueue.m.Unlock()
		}
	}()

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

	wg.Wait()
	close(chanRequest)
	close(chanResponse)
}
