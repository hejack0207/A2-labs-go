/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

// go.mod >>>
// module github.com/gorun/autossh
// go 1.13
// require github.com/ThomasRooney/gexpect latest
// require github.com/spf13/cobra v1.2.1
// <<< go.mod
//
// go.env >>>
// GO111MODULE=on
// <<< go.env

package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ThomasRooney/gexpect"
	"github.com/spf13/cobra"
)

type jumpserver struct {
	user     string
	password string
	host     string
	port     string
}

func main() {
	cmd := &cobra.Command{
		Use:   "autossh",
		Short: "auto start interactive ssh session",
		// Long: `auto start interactive ssh session,
		//         user password can be specified as command line options,
		//         jump server is optional`,
	}
	sjumpserver := cmd.Flags().StringP("jumpserver", "j", "", "jumpserver like: user/passwd@host:port")

	regex := regexp.MustCompile("(.*)/(.*)@(.*):(.*)")
	matches := regex.FindAllStringSubmatch(*sjumpserver, 1)

	jserver := jumpserver{}
	if len(matches) > 0 {
		jserver.user = matches[0][1]
		jserver.password = matches[0][2]
		jserver.host = matches[0][3]
		jserver.port = matches[0][4]
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("sjumpserver: %s\n", *sjumpserver)
		fmt.Printf("jumpserver: %s\n", jserver)
		if *sjumpserver == "" {
			os.Exit(1)
		}
		process, _ := gexpect.Spawn("sh")
		process.Interact()
		process.Close()
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
