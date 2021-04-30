package utils

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

const NUM_OF_TEST_CASES = 65535

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

func TestPush(t *testing.T) {
	defer func() {
		var r = recover()
		if r != nil {
			t.Error(r)
		}
	}()

	var queue = new(Queue)
	queue.Init()

	var values [NUM_OF_TEST_CASES]int
	for i := 0; i < NUM_OF_TEST_CASES; i++ {
		values[i] = rand.Int()
	}

	for _, v := range values {
		queue.push(v)
	}

	var e = queue.Front()
	for i := 0; i < queue.Len(); i++ {
		if values[i] != e.Value {
			t.Errorf("%d != %d", values[i], e.Value)
			break
		}
		e = e.Next()
	}
}

func TestPop(t *testing.T) {
	defer func() {
		var r = recover()
		if r != nil {
			t.Error(r)
		}
	}()

	var queue = new(Queue)
	queue.Init()

	for i := 0; i < NUM_OF_TEST_CASES; i++ {
		queue.push(rand.Int())
	}

	for i := 0; i < NUM_OF_TEST_CASES; i++ {
		queue.pop()
	}

	if queue.Len() != 0 {
		t.Error("Queue is not emptied.")
	}
}
