package orderedmap_test

import (
	"fmt"

	"github.com/wk8/go-ordered-map"
)

func Example() {
	om := orderedmap.New()

	om.Set("foo", "bar")
	om.Set("bar", "baz")
	om.Set("coucou", "toi")

	fmt.Println("## Get operations: ##")
	fmt.Println(om.Get("foo"))
	fmt.Println(om.Get("i dont exist"))

	fmt.Println("## Iterating over pairs from oldest to newest: ##")
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		fmt.Printf("%s => %s\n", pair.Key, pair.Value)
	}

	fmt.Println("## Iterating over the 2 newest pairs: ##")
	i := 0
	for pair := om.Newest(); pair != nil; pair = pair.Prev() {
		fmt.Printf("%s => %s\n", pair.Key, pair.Value)
		i++
		if i >= 2 {
			break
		}
	}

	// Output:
	// ## Get operations: ##
	// bar true
	// <nil> false
	// ## Iterating over pairs from oldest to newest: ##
	// foo => bar
	// bar => baz
	// coucou => toi
	// ## Iterating over the 2 newest pairs: ##
	// coucou => toi
	// bar => baz
}
