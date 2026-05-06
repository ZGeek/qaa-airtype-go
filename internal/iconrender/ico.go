package iconrender

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/png"
)

// WindowsICOSizes 覆盖任务栏/资源管理器常用 DPI；单张 32×32 放大到 256 会糊。
var WindowsICOSizes = []int{16, 24, 32, 48, 64, 128, 256}

// EncodeWindowsICO 生成内含多枚 PNG 的 ICO（Vista+ 通用），用于 exe 与托盘。
func EncodeWindowsICO() ([]byte, error) {
	sizes := WindowsICOSizes
	pngs := make([][]byte, len(sizes))
	for i, sz := range sizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, Render(sz)); err != nil {
			return nil, err
		}
		pngs[i] = buf.Bytes()
	}
	return buildPNGPackedICO(sizes, pngs)
}

// buildPNGPackedICO 将多枚 PNG 按顺序写入 ICO；目录项顺序与数据块顺序一致。
func buildPNGPackedICO(sizes []int, pngs [][]byte) ([]byte, error) {
	n := len(sizes)
	if n != len(pngs) {
		return nil, fmt.Errorf("iconrender: sizes/pngs 长度不一致")
	}
	if n == 0 {
		return nil, fmt.Errorf("iconrender: 空 ICO")
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
