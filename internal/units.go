package internal

import "fmt"

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// ByteSize represents a byte quantity
type ByteSize int64

// String gives a human readable string
// Based on https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func (b ByteSize) String() string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c",
		float64(b)/float64(div), "KMGTPE"[exp])
}
