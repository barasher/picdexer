package dispatch

import (
	"context"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestDispatchTasks(t *testing.T) {
	in := make(chan browse.Task, 2)
	out1 := make(chan browse.Task, 2)
	out2 := make(chan browse.Task, 2)

	in <- browse.Task{Path: "p1"}
	in <- browse.Task{Path: "p2"}
	close(in)

	wg := sync.WaitGroup{}
	wg.Add(2)

	dispatched1 := []string{}
	go func() {
		for cur := range out1 {
			dispatched1 = append(dispatched1, cur.Path)
		}
		wg.Done()
	}()

	dispatched2 := []string{}
	go func() {
		for cur := range out2 {
			dispatched2 = append(dispatched2, cur.Path)
		}
		wg.Done()
	}()

	DispatchTasks(context.TODO(), in, out1, out2)

	wg.Wait()
	assert.Equal(t, []string{"p1", "p2"}, dispatched1)
	assert.Equal(t, []string{"p1", "p2"}, dispatched2)
}
