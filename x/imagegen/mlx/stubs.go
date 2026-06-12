//go:build !cgo

package mlx

import (
	"errors"
)

type Array struct{}

func (a *Array) Dtype() Dtype { return 0 }
func (a *Array) Shape() []int32 { return nil }
func (a *Array) Data() []float32 { return nil }
func (a *Array) DataInt32() []int32 { return nil }
func (a *Array) DataFloat32() []float32 { return nil }
func (a *Array) Nbytes() int64 { return 0 }
func (a *Array) Valid() bool { return false }
func (a *Array) Kept() bool { return false }
func (a *Array) Eval() *Array { return a }
func (a *Array) Free() {}
func (a *Array) Ndim() int { return 0 }
func (a *Array) Dim(axis int) int { return 0 }

type Dtype int

const (
	DtypeBool Dtype = iota
	DtypeUint8
	DtypeUint16
	DtypeUint32
	DtypeUint64
	DtypeInt8
	DtypeInt16
	DtypeInt32
	DtypeInt64
	DtypeFloat16
	DtypeFloat32
	DtypeFloat64
	DtypeBFloat16
	DtypeComplex64
)

type Stream struct{}

type SafetensorsFile struct{}

func (s *SafetensorsFile) Free() {}
func (s *SafetensorsFile) GetMetadata(key string) string { return "" }
func (s *SafetensorsFile) Get(name string) *Array { return nil }
func (s *SafetensorsFile) Set(name string, arr *Array) {}

func LoadSafetensorsNative(path string) (*SafetensorsFile, error) {
	return nil, errors.New("MLX requires CGO")
}

func Add(a, b *Array) *Array { return nil }
func AddMM(c, a, b *Array, alpha, beta float32) *Array { return nil }
func AddScalar(a *Array, s float32) *Array { return nil }
func ArangeInt(start, stop, step int32, dtype Dtype) *Array { return nil }
func Argmax(a *Array, axis int, keepdims bool) *Array { return nil }
func Argpartition(a *Array, kth int, axis int) *Array { return nil }
func Argsort(a *Array, axis int) *Array { return nil }
func AsType(a *Array, dtype Dtype) *Array { return nil }
func AsyncEval(outputs ...*Array) {}
func BroadcastTo(a *Array, shape []int32) *Array { return nil }
func ClearCache() {}
func ClipScalar(a *Array, minVal, maxVal float32, hasMin, hasMax bool) *Array { return nil }
func Collect(v any) []*Array { return nil }
func Concatenate(arrays []*Array, axis int) *Array { return nil }
func Contiguous(a *Array) *Array { return nil }
func Conv2d(input, weight *Array, stride, padding int32) *Array { return nil }
func Cos(a *Array) *Array { return nil }
func Cumsum(a *Array, axis int) *Array { return nil }
func Dequantize(w, scales, biases *Array, groupSize, bits int, mode string) *Array { return nil }
func Div(a, b *Array) *Array { return nil }
func DivScalar(a *Array, s float32) *Array { return nil }
func EnableCompile() {}
func Eval(outputs ...*Array) []*Array { return nil }
func ExpandDims(a *Array, axis int) *Array { return nil }
func Full(value float32, shape ...int32) *Array { return nil }
func FullDtype(value float32, dtype Dtype, shape ...int32) *Array { return nil }
func GPUIsAvailable() bool { return false }
func GetDefaultStream() *Stream { return nil }
func GetMemoryLimit() uint64 { return 0 }
func InitMLX() error { return nil }
func Keep(arrays ...*Array) {}
func LayerNorm(x *Array, eps float32) *Array { return nil }
func LessScalar(a *Array, s float32) *Array { return nil }
func Linear(a, weight *Array) *Array { return nil }
func Matmul(a, b *Array) *Array { return nil }
func Mean(a *Array, axis int, keepdims bool) *Array { return nil }
func MetalGetActiveMemory() uint64 { return 0 }
func MetalGetPeakMemory() uint64 { return 0 }
func MetalIsAvailable() bool { return false }
func MetalResetPeakMemory() {}
func MetalSetWiredLimit(limit uint64) uint64 { return 0 }
func MetalStartCapture(path string) {}
func MetalStopCapture() {}
func Mul(a, b *Array) *Array { return nil }
func MulScalar(a *Array, s float32) *Array { return nil }
func Neg(a *Array) *Array { return nil }
func NewArray(data []float32, shape []int32) *Array { return nil }
func NewArrayFloat32(data []float32, shape []int32) *Array { return nil }
func NewArrayInt32(data []int32, shape []int32) *Array { return nil }
func NewScalarArray(value float32) *Array { return nil }
func NewStream() *Stream { return nil }
func Ones(shape ...int32) *Array { return nil }
func Pad(a *Array, paddings []int32) *Array { return nil }
func Quantize(w *Array, groupSize, bits int, mode string) (weights, scales, biases *Array) { return nil, nil, nil }
func QuantizedMatmul(x, w, scales, biases *Array, transpose bool, groupSize, bits int, mode string) *Array { return nil }
func RMSNorm(x, weight *Array, eps float32) *Array { return nil }
func RSqrt(a *Array) *Array { return nil }
func RandomCategorical(logits *Array, axis int, numSamples int) *Array { return nil }
func RandomNormal(shape []int32, seed uint64) *Array { return nil }
func RandomNormalWithDtype(shape []int32, seed uint64, dtype Dtype) *Array { return nil }
func Reshape(a *Array, shape ...int32) *Array { return nil }
func RestoreDefaultErrorHandler() {}
func ScaledDotProductAttention(q, k, v *Array, scale float32, causalMask bool) *Array { return nil }
func ScaledDotProductAttentionWithSinks(q, k, v *Array, scale float32, maskMode string, mask, sinks *Array) *Array { return nil }
func SetDefaultDeviceGPU() {}
func SetDefaultStream(s *Stream) {}
func SiLU(a *Array) *Array { return nil }
func Sin(a *Array) *Array { return nil }
func Slice(a *Array, start, stop []int32) *Array { return nil }
func SliceUpdateInplace(a, update *Array, start, stop []int32) *Array { return nil }
func Softmax(a *Array, axis int) *Array { return nil }
func Sqrt(a *Array) *Array { return nil }
func Square(a *Array) *Array { return nil }
func Squeeze(a *Array, axis int) *Array { return nil }
func Sub(a, b *Array) *Array { return nil }
func Take(a *Array, indices *Array, axis int) *Array { return nil }
func TakeAlongAxis(a, indices *Array, axis int) *Array { return nil }
func Tanh(a *Array) *Array { return nil }
func Tile(a *Array, reps []int32) *Array { return nil }
func ToBFloat16(a *Array) *Array { return nil }
func Transpose(a *Array, axes ...int) *Array { return nil }
func Tri(n, m int32, k int) *Array { return nil }
func Where(condition, a, b *Array) *Array { return nil }
func Zeros(shape []int32, dtype ...Dtype) *Array { return nil }
