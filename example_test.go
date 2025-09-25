package hashcache

import (
	"context"
	"fmt"

	"github.com/gford1000-go/hasher"
)

func Example() {

	// Create the cache, using sha512 hashing for the key
	cache, _ := New[string](context.Background(), WithHashType(hasher.Sha512))
	defer cache.Delete() // Tidy up

	data := "Hellow World"

	// Store some data
	cache.Put("MyKey", data)

	// Retrieve data
	resp, _ := cache.Get("MyKey")

	fmt.Println(resp == data)
	// Output: true
}
