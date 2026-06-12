//go:build !cgo

// Package mlx provides stubs for non-cgo builds.
package mlx

import (
	"errors"
	"fmt"
	"iter"
)

type Array struct{}

func (t *Array) DType() DType { return 0 }
func (t *Array) Dtype() Dtype { return 0 }
func (t *Array) Shape() []int32 { return nil }
func (t *Array) Dims() []int { return nil }
func (t *Array) NumDims() int { return 0 }
func (t *Array) DataFloat32() []float32 { return nil }
func (t *Array) Valid() bool { return false }
func (t *Array) Free() {}

func (t *Array) Abs() *Array { return t }
func (t *Array) Add(other *Array) *Array { return t }
func (t *Array) AddScalar(s float32) *Array { return t }
func (t *Array) Addmm(a, b *Array, alpha, beta float32) *Array { return t }
func (t *Array) Argmax(axis int, keepDims bool) *Array { return t }
func (t *Array) ArgpartitionAxis(kth int, axis int) *Array { return t }
func (t *Array) ArgsortAxis(axis int) *Array { return t }
func (t *Array) AsType(dtype DType) *Array { return t }
func (t *Array) CategoricalWithKey(axis int, key *Array) *Array { return t }
func (t *Array) Concatenate(axis int, others ...*Array) *Array { return t }
func (t *Array) Cumsum(axis int, reverse, inclusive bool) *Array { return t }
func (t *Array) Divide(other *Array) *Array { return t }
func (t *Array) Equal(other *Array) *Array { return t }
func (t *Array) ExpandDims(axis int) *Array { return t }
func (t *Array) Greater(other *Array) *Array { return t }
func (t *Array) Less(other *Array) *Array { return t }
func (t *Array) LessEqual(other *Array) *Array { return t }
func (t *Array) Multiply(other *Array) *Array { return t }
func (t *Array) Negative() *Array { return t }
func (t *Array) PutAlongAxis(indices, values *Array, axis int) *Array { return t }
func (t *Array) Reshape(axes ...int) *Array { return t }
func (t *Array) ScatterAddAxis(indices, values *Array, axis int) *Array { return t }
func (t *Array) Slice(slices ...slice) *Array { return t }
func (t *Array) Squeeze(axis int) *Array { return t }
func (t *Array) Subtract(other *Array) *Array { return t }
func (t *Array) SumAxis(axis int, keepDims bool) *Array { return t }
func (t *Array) TakeAlongAxis(indices *Array, axis int) *Array { return t }
func (t *Array) TakeAxis(indices *Array, axis int) *Array { return t }
func (t *Array) Transpose(axes ...int) *Array { return t }
func (t *Array) GatherMM(other, lhs, rhs *Array, sorted bool) *Array { return t }
func (t *Array) Matmul(other *Array) *Array { return t }
func (t *Array) Sigmoid() *Array { return t }
func (t *Array) Tanh() *Array { return t }
func (t *Array) Set(other *Array) {}
func (t *Array) LogsumexpAxis(axis int, keepDims bool) *Array { return t }
func (t *Array) MaxAxis(axis int, keepDims bool) *Array { return t }
func (t *Array) Clone() *Array { return t }
func (t *Array) NumBytes() int { return 0 }
func (t *Array) SliceUpdate(other *Array, slices ...slice) *Array { return t }

// Additional methods for convenience/safety
func (t *Array) Dim(axis int) int { return 0 }
func (t *Array) MultiplyScalar(s float32) *Array { return t }
func (t *Array) Mul(other *Array) *Array { return t }
func (t *Array) Sub(other *Array) *Array { return t }

type DType int

const (
	DTypeBool DType = iota
	DTypeUint8
	DTypeUint16
	DTypeUint32
	DTypeUint64
	DTypeInt8
	DTypeInt16
	DTypeInt32
	DTypeInt64
	DTypeFloat16
	DTypeFloat32
	DTypeFloat64
	DTypeBFloat16
	DTypeComplex64
)

// For backward compatibility / safety
type Dtype = DType
const (
	DtypeFloat32 = DTypeFloat32
	DtypeBFloat16 = DTypeBFloat16
)

type SafetensorsFile struct{}

func (s *SafetensorsFile) Free() {}
func (s *SafetensorsFile) GetMetadata(key string) string { return "" }
func (s *SafetensorsFile) Get(name string) *Array { return nil }

func LoadSafetensorsNative(path string) (*SafetensorsFile, error) {
	return nil, errors.New("MLX requires CGO")
}

func AsType(arr *Array, dtype DType) *Array { return arr }
func Contiguous(a *Array, allowColMajor bool) *Array { return a }
func Eval(arrs ...*Array) {}
func AsyncEval(outputs ...*Array) {}
func NewArray(data []float32, shape []int32) *Array { return nil }
func NewArrayInt32(data []int32, shape []int32) *Array { return nil }
func ClipScalar(arr *Array, min, max float32) *Array { return arr }
func Transpose(arr *Array, axes ...int) *Array { return arr }

func GPUIsAvailable() bool { return false }
func SetDefaultDeviceGPU() {}
func EnableCompile() {}
func Collect(v any) []*Array { return nil }
func ToBFloat16(arr *Array) *Array { return arr }
func MetalGetActiveMemory() uint64 { return 0 }
func MetalResetPeakMemory() {}
func MetalStartCapture(path string) {}
func MetalIsAvailable() bool { return false }
func Pin(s ...*Array) {}
func Sweep() {}
func Unpin(s ...*Array) {}

// Range-over-func load
func Load(path string) iter.Seq2[string, *Array] { return nil }

func ActiveMemory() int { return 0 }
func AddScalar(a *Array, s float32) *Array { return nil }
func BernoulliWithKey(p *Array, key *Array) *Array { return nil }
func CacheMemory() int { return 0 }
func CheckInit() error { return nil }
func Concatenate(arrays []*Array, axis int) *Array { return nil }
func DivScalar(a *Array, s float32) *Array { return nil }
func FastScaledDotProductAttention(q, k, v *Array, scale float32, mode string, mask *Array) *Array { return nil }
func Log(a *Array) *Array { return nil }
func Maximum(a, b *Array) *Array { return nil }
func Minimum(a, b *Array) *Array { return nil }
func MulScalar(a *Array, s float32) *Array { return nil }
func PeakMemory() int { return 0 }
func PrettyBytes(n int) fmt.Stringer { return nil }
func Dequantize(w, scales, biases *Array, groupSize, bits int, mode string) *Array { return nil }
func QuantizedMatmul(x, w, scales, biases *Array, transpose bool, groupSize, bits int, mode string) *Array { return nil }
func Tri(n, m int32, k int) *Array { return nil }
func Mul(a, b *Array) *Array { return nil }
func LayerNormFn(x, weight, bias *Array, eps float32) *Array { return nil }
func RMSNormFn(x, weight *Array, eps float32) *Array { return nil }
func SliceStartStop(a *Array, start, stop []int32) *Array { return nil }
func TakeAlongAxis(a, indices *Array, axis int) *Array { return nil }
func ExpandDims(a *Array, axis int) *Array { return nil }
func NewScalarArray(value float32) *Array { return nil }
func Quantize(w *Array, groupSize, bits int, mode string) (weights, scales, biases *Array) { return nil, nil, nil }
func RandomKey(seed uint64) *Array { return nil }
func SoftmaxAxis(a *Array, axis int, precise bool) *Array { return nil }
func Tile(a *Array, reps []int32) *Array { return nil }
func Version() string { return "" }
func Where(condition, a, b *Array) *Array { return nil }
func Zeros(dtype DType, shape ...int) *Array { return nil }

func FromValue[T any](t T) *Array { return nil }
func FromValues[T any](s []T, shape ...int) *Array { return nil }

type slice struct {
	args []int
}

const End = 2147483647 // math.MaxInt32

func Slice(args ...int) slice {
	return slice{args: args}
}

// Fused kernel compilation types and options
type CompileFunc func(inputs ...*Array) []*Array
type CompileOption func(*compileConfig)
type compileConfig struct {
	shapeless bool
}

func Shapeless() CompileOption {
	return func(c *compileConfig) {}
}

func Compile(name string, fn CompileFunc, opts ...CompileOption) CompileFunc {
	return fn
}

func Compile1(name string, fn func(*Array) *Array, opts ...CompileOption) func(*Array) *Array {
	return fn
}

func Compile2(name string, fn func(*Array, *Array) *Array, opts ...CompileOption) func(*Array, *Array) *Array {
	return fn
}

func Compile3(name string, fn func(*Array, *Array, *Array) *Array, opts ...CompileOption) func(*Array, *Array, *Array) *Array {
	return fn
}

func Logaddexp(a, b *Array) *Array { return a }
func Conv1d(x, weight *Array, bias *Array, stride, padding, dilation, groups int32) *Array { return nil }
func Add(a, b *Array) *Array { return nil }
func FastGatedDelta(q, k, v, g, beta, state, mask *Array) (y, nextState *Array) { return nil, nil }
func Reshape(a *Array, shape ...int32) *Array { return nil }
func RoPEWithFreqs(x *Array, dims int, traditional bool, base, scale float32, offsets *Array, freqs *Array) *Array { return nil }
func RoPEWithBase(x *Array, dims int, traditional bool, base, scale float32, offsets *Array) *Array { return nil }
func Matmul(a, b *Array) *Array { return nil }
func Squeeze(a *Array, axis int) *Array { return nil }
func Stack(arrays []*Array, axis int) *Array { return nil }
func GatherQMM(x, w, scales *Array, biases, lhsIndices, rhsIndices *Array, transpose bool, groupSize, bits int, mode string, sortedIndices bool) *Array { return nil }
func Take(a *Array, indices *Array, axis int) *Array { return nil }
func FloorDivideScalar(a *Array, s int32) *Array { return nil }
func Argsort(a *Array, axis int) *Array { return nil }
func Flatten(a *Array) *Array { return nil }
func Argpartition(a *Array, kth int, axis int) *Array { return nil }
func Neg(a *Array) *Array { return nil }
