//go:build windows

package main

import (
	"github.com/QAA-Tools/qaa-airtype/go/internal/iconrender"
)

func generateIcon() []byte {
	data, err := iconrender.EncodeWindowsICO()
	if err != nil {
		return nil
	}
	return data
}