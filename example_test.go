package hashcache

import (
	"context"
	"fmt"
)

func Example() {

	// Create the cache
	cache, _ := New(context.Background(), nil)
	defer func() {
		cache.Delete() // Tidy up
	}()

	// Store some data
	cache.Put("MyKey", "MyValue")

	// Retrieve data
	resp, _ := cache.Get("MyKey")

	fmt.Println(resp)
	// Output: MyValue
}
