package version

import "fmt"

func Exec(flags ...string) bool {
	if cantContinue(flags) {
		return false
	}

	ver := "v0.0.1"
	name := "Bareborn BB Command"

	fmt.Println(name)
	fmt.Println("version: " + ver)

	return true
}

// cantContinue コマンドが続けられるか判定する
func cantContinue(flags []string) bool {
	if len(flags) != 0 && flags[0] == "help" {
		cmdDetail()
		return true
	}
	return false
}

// cmdDetail コマンド詳細を表示する
func cmdDetail() {
	fmt.Print("バージョン情報を表示します。")
}
