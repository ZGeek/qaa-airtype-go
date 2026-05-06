// Package iconrender 从内嵌 PNG 生成多尺寸 ICO，用于 exe 图标和系统托盘。
package iconrender

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
)

//go:embed icon.png
var embeddedIconPNG []byte

// Render 将内嵌 PNG 缩放为指定尺寸的 RGBA 图像。
func Render(size int) *image.RGBA {
	if size < 16 {
		size = 16
	}

	src, err := png.Decode(bytes.NewReader(embeddedIconPNG))
	if err != nil {
		// 解码失败则返回纯色兜底
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		fillSolid(img, size)
		return img
	}

	return resizeRGBA(src, size)
}

// resizeRGBA 将任意 image.Image 缩放到 size×size 的 RGBA。
func resizeRGBA(src image.Image, size int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			srcX := x*sw/size + sb.Min.X
			srcY := y*sh/size + sb.Min.Y
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func fillSolid(img *image.RGBA, size int) {
	c := color.RGBA{R: 26, G: 171, B: 168, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, c)
		}
	}
}