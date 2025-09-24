package hashcache

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	irand "math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

type testResult struct {
	testerID   int
	err        error
	putElapsed int64
	getElapsed int64
}

func tester(testerID int, c *Cache, addItemsCount int, getCount int, resultChan chan *testResult) {

	createRandom := func(size int) ([]byte, error) {
		values := make([]byte, size)
		if _, err := io.ReadFull(rand.Reader, values); err != nil {
			return nil, fmt.Errorf("Error creating random sequence: %s", err)
		}
		return values, nil
	}

	// 1. Create the specified number of random entries in Cache

	type entry struct {
		key   any
		value any
	}

	putStart := time.Now()

	entries := make([]*entry, 0)
	keys := make([]any, 0)

	for range addItemsCount {
		key, _ := createRandom(64)
		value, _ := createRandom(1 + irand.Intn(1024))

		e := &entry{key: key, value: value}
		entries = append(entries, e)
		keys = append(keys, key)

		c.Put(key, value)
	}

	putEnd := time.Now()

	// 2. For each entry, test retrieval from Cache getCount times
	//    Stop if any retrieval returns an error, exit test for that entry

	errChan := make(chan error, addItemsCount*getCount)

	var wg sync.WaitGroup
	wg.Add(addItemsCount * getCount)

	for i := range addItemsCount {
		go func(myI int) {
			for gets := range getCount {
				go func(myGets int) {
					defer func() {
						wg.Done()
					}()

					resp, err := c.Get(keys[myI])

					if err != nil {
						errChan <- fmt.Errorf("After %d cycles, Key[%d] had error: %v", myGets, myI, err)
						return
					}

					if !reflect.DeepEqual(resp, entries[myI].value) {
						errChan <- fmt.Errorf("After %d cycles, Key[%d] had value mismatch", myGets, myI)
						return
					}
				}(gets)
			}

		}(i)
	}

	getEnd := time.Now()

	wg.Wait()

	// 3. Check for errors and return result

	result := testResult{
		testerID:   testerID,
		putElapsed: putEnd.Sub(putStart).Nanoseconds(),
		getElapsed: getEnd.Sub(putEnd).Nanoseconds(),
	}

	if len(errChan) > 0 {
		result.err = <-errChan
	}

	resultChan <- &result
}

func TestStress(t *testing.T) {

	cache, _ := New(context.Background(), &Config{RequestBuffer: 10000})

	testerCount := 2    // How many testers will be started
	itemsInCache := 200 // How many items each tester() will add to the cache
	retrieveCount := 20 // How many retrievals are attempted of each item

	resultChan := make(chan *testResult, testerCount)

	var wg sync.WaitGroup

	for i := range testerCount {
		wg.Add(1)
		go func(myI int) {
			defer func() {
				wg.Done()
			}()

			tester(myI, cache, itemsInCache, retrieveCount, resultChan)
		}(i)
	}

	wg.Wait()

	for len(resultChan) > 0 {
		result := <-resultChan
		if result.err != nil {
			t.Errorf("Unexpected error: %v", result.err)
			return
		}
		fmt.Printf("Get: %10d, Put: %10d\n", result.getElapsed, result.putElapsed)
	}
}
