package version

import "fmt"

func Exec(flags ...string) {
	ver := "v0.0.1"
	name := "Bareborn BB Command"

	fmt.Println(name)
	fmt.Println("version: " + ver)
}
