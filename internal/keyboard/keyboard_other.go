//go:build !windows

package keyboard

func Paste() error {
	return nil
}