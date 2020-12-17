package internal

import "fmt"

const defaultBinaryManagerThreadCount = 4

type BinaryManager struct {
	threadCount int
}

func NewBinaryManager(opts ...func(*BinaryManager)error) (*BinaryManager, error){
	bm := &BinaryManager{threadCount:defaultBinaryManagerThreadCount}
	for _, cur := range opts {
		if err := cur(bm) ; err != nil {
			return nil, fmt.Errorf("error while creating EsPusher: %w", err)
		}
	}
	return bm, nil
}

func BinaryManagerThreadCount(c int) func(*BinaryManager) error {
	return func(bm *BinaryManager) error {
		if c <= 0 {
			return fmt.Errorf("wrong thread count value (%v), must be > 0", c)
		}
		bm.threadCount = c
		return nil
	}
}