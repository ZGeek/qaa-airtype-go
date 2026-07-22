//go:build !windows

package keyboard

func Paste() error {
	return nil
}

func Enter() error {
	return nil
}

func TypeText(text string) error {
	return nil
}

func StartTouchScroll() error {
	return nil
}

func MoveTouchScroll(offsetY float64) error {
	return nil
}

func EndTouchScroll() error {
	return nil
}

func MoveMouse(dx float64, dy float64) error {
	return nil
}

func LeftClick() error {
	return nil
}

func LeftDown() error {
	return nil
}

func LeftUp() error {
	return nil
}

func RightClick() error {
	return nil
}
