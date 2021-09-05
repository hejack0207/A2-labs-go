/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

// go.mod >>>
// module github.com/gorun/graph-split
// go 1.13
// require github.com/spf13/cobra v1.2.1
// <<< go.mod
//
// go.env >>>
// GO111MODULE=on
// <<< go.env

package main

func main() {
	println("hi,gorun!")
}
