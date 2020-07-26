// +build !linux
// +build !darwin
// +build !freebsd
// +build !nacl
// +build !windows

package progressbar

func getWidth() int {
	return 80
}
