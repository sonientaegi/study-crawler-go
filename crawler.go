package main

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
	"study-crawler-go/utils"
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

		var url string
		var isChildOfLink bool

		var element = queue.Pop().(elementType)
		if element.node.Type == html.ElementNode {
			switch element.node.Data {
			case "a":

			case "div":

			}
		}
	}

	return userNames
}

func main() {

}
