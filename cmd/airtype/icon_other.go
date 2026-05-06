//go:build !windows

package main

import (
	"image"
	
	"github.com/QAA-Tools/qaa-airtype/go/internal/iconrender"
)

func generateIcon() []byte {
	return nil
}

func generateIconImage() *image.RGBA {
	return iconrender.Render(32)
}