package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestDispatchTasks(t *testing.T) {
	in := make(chan Task, 2)
	out1 := make(chan Task, 2)

	in <- Task{		Path: "p1"	}
	in <- Task{		Path: "p2"	}
	close(in)

	wg := sync.WaitGroup{}
	wg.Add(1)
	dispatched := []string{}
	go func(){
		for cur := range out1 {
			dispatched = append(dispatched, cur.Path)
		}
		wg.Done()
	}()

	DispatchTasks(context.TODO(), in, out1)

	wg.Wait()
	assert.Equal(t, []string{"p1", "p2"}, dispatched)
}