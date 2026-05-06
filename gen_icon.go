//go:build ignore

package main

import (
	"os"

	"github.com/QAA-Tools/qaa-airtype/go/internal/iconrender"
)

func main() {
	data, err := iconrender.EncodeWindowsICO()
	if err != nil {
		panic(err)
	}
	
	err = os.WriteFile("cmd/airtype/app.ico", data, 0644)
	if err != nil {
		panic(err)
	}
	
	println("Generated cmd/airtype/app.ico")
}