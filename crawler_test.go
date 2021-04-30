package main

import (
	"fmt"
	"testing"
)

const TEST_URL = "https://github.com/SonienTaegi?tab=followers"

func TestFetch(t *testing.T) {
	node := fetch(TEST_URL)
	if node == nil {
		t.Error("Can not retrieve URL", TEST_URL)
	}
	fmt.Println(node)
}

func TestParseFollow(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	var node = fetch(TEST_URL)
	var userNames = parseFollow(node)

	fmt.Println(userNames)
}
