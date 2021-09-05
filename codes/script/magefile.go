// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func Help() {
	fmt.Print("help invoked!")
	meminfo, _ := ioutil.ReadFile("/proc/meminfo")
	s_meminfo := string(meminfo)
	meminfo_lines := strings.Split(s_meminfo, "\n")
	fmt.Print(s_meminfo)
	fmt.Print(len(meminfo_lines))
}

