package main

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	var x = []rune("aaa")
	fmt.Println(len(x))
}

func Test2(t *testing.T) {
	t.Log("hi")
}
