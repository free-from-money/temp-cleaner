//go:build !((linux || darwin || windows) && (amd64 || arm64))

package tempcleaner

var binGz []byte
