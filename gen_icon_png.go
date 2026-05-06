//go:build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
)

// WindowsICOSizes 覆盖任务栏/资源管理器常用 DPI
var WindowsICOSizes = []int{16, 24, 32, 48, 64, 128, 256}

func main() {
	// 读取 PNG 文件
	pngFile := "C:/Users/ZG/Downloads/voice.png"
	input, err := os.ReadFile(pngFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to read PNG: %v", err))
	}

	// 解码 PNG
	img, err := png.Decode(bytes.NewReader(input))
	if err != nil {
		panic(fmt.Sprintf("Failed to decode PNG: %v", err))
	}

	// 生成多尺寸 ICO
	sizes := WindowsICOSizes
	pngs := make([][]byte, len(sizes))
	
	for i, sz := range sizes {
		// 缩放图像到指定尺寸
		resized := resizeImage(img, sz)
		
		// 编码为 PNG
		var buf bytes.Buffer
		if err := png.Encode(&buf, resized); err != nil {
			panic(fmt.Sprintf("Failed to encode PNG for size %d: %v", sz, err))
		}
		pngs[i] = buf.Bytes()
	}

	// 构建 ICO 文件
	icoData, err := buildPNGPackedICO(sizes, pngs)
	if err != nil {
		panic(fmt.Sprintf("Failed to build ICO: %v", err))
	}

	// 写入 ICO 文件
	outputPath := "cmd/airtype/app.ico"
	if err := os.WriteFile(outputPath, icoData, 0644); err != nil {
		panic(fmt.Sprintf("Failed to write ICO: %v", err))
	}

	fmt.Println("Generated cmd/airtype/app.ico from PNG")
}

func resizeImage(src image.Image, size int) image.Image {
	// 使用简单的缩放算法
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// 计算源图像中的对应位置
			srcX := (x * srcW) / size
			srcY := (y * srcH) / size
			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}

func buildPNGPackedICO(sizes []int, pngs [][]byte) ([]byte, error) {
	n := len(sizes)
	if n != len(pngs) {
		return nil, fmt.Errorf("sizes/pngs length mismatch")
	}
	if n == 0 {
		return nil, fmt.Errorf("empty ICO")
	}

	headerSize := 6 + 16*n
	offset := uint32(headerSize)

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // type: icon
	_ = binary.Write(&buf, binary.LittleEndian, uint16(n))

	for i := 0; i < n; i++ {
		wb, hb := icoWHBytes(sizes[i], sizes[i])
		buf.WriteByte(wb)
		buf.WriteByte(hb)
		buf.WriteByte(0) // color count
		buf.WriteByte(0) // reserved
		_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // planes
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // bpp: 0 = PNG
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(pngs[i])))
		_ = binary.Write(&buf, binary.LittleEndian, offset)
		offset += uint32(len(pngs[i]))
	}

	for i := 0; i < n; i++ {
		buf.Write(pngs[i])
	}

	return buf.Bytes(), nil
}

func icoWHBytes(w, h int) (byte, byte) {
	if w > 256 {
		w = 256
	}
	if h > 256 {
		h = 256
	}
	var wb, hb byte
	if w >= 256 {
		wb = 0
	} else {
		wb = byte(w)
	}
	if h >= 256 {
		hb = 0
	} else {
		hb = byte(h)
	}
	return wb, hb
}