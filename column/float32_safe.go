//go:build !(386 || amd64 || amd64p32 || arm || arm64 || mipsle || mips64le || mips64p32le || ppc64le || riscv || riscv64) || purego
// +build !386,!amd64,!amd64p32,!arm,!arm64,!mipsle,!mips64le,!mips64p32le,!ppc64le,!riscv,!riscv64 purego

package column

import (
	"encoding/binary"
	"math"
)

// ReadAll read all value in this block and append to the input slice
func (c *Float32) ReadAll(value *[]float32) {
	for i := 0; i < c.totalByte; i += c.size {
		*value = append(*value,
			math.Float32frombits(binary.LittleEndian.Uint32(c.b[i:i+c.size])))
	}
}