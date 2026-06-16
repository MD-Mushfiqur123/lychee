//go:build !darwin

package client

import (
	"fmt"
	"io"

	"github.com/MD-Mushfiqur123/lychee/x/create"
)

// QuantizeSupported returns false on non-darwin platforms (quantization requires MLX).
func QuantizeSupported() bool {
	return false
}

func quantizeTensor(r io.Reader, name, dtype string, shape []int32, quantize string) ([]byte, error) {
	return nil, fmt.Errorf("quantization is not supported on this platform (requires MLX/darwin)")
}

func quantizePackedGroup(groupName string, tensors []create.PackedTensorInput) ([]byte, error) {
	return nil, fmt.Errorf("quantization is not supported on this platform (requires MLX/darwin)")
}
