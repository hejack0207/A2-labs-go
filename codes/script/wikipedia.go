/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

// go.mod >>>
// module github.com/gorun/script
// go 1.13
// require github.com/spf13/cobra v1.2.1
// require github.com/ishaangandhi/wikigopher latest
// require github.com/ogier/pflag latest
// <<< go.mod
//
// go.env >>>
// GO111MODULE=on
// <<< go.env

package main

import (
	"fmt"
	"log"

	"github.com/ishaangandhi/wikigopher"
	"github.com/ogier/pflag"
)

func main() {
	pflag.Parse()
	terms := pflag.Args()
	for _, term := range terms {
		log.Printf("term: %s", term)
		page, err := wikigopher.Page(term)
		if err != nil {
			log.Printf("error when fetch term: %s", term)
		}
		fmt.Println(page.Content)
	}
}
