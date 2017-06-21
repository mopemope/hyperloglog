package hlltc

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func estimateError(got, exp uint64) float64 {
	var delta uint64
	if got > exp {
		delta = got - exp
	} else {
		delta = exp - got
	}
	return float64(delta) / float64(exp)
}

func TestCardinality(t *testing.T) {
	hlltc, err := New(14)
	if err != nil {
		t.Error("expected no error, got", err)
	}

	step := 10
	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := fmt.Sprintf("flow-%d", i)
		hlltc.Insert([]byte(str))
		unique[str] = true

		if len(unique)%step == 0 {
			step *= 5
			exact := uint64(len(unique))
			res := uint64(hlltc.Estimate())
			ratio := 100 * estimateError(res, exact)
			if ratio > 2 {
				t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
			}
		}
	}
	exact := uint64(len(unique))
	res := uint64(hlltc.Estimate())
	ratio := 100 * estimateError(res, exact)
	if ratio > 2 {
		t.Errorf("Exact %d, got %d which is %.2f%% error", exact, res, ratio)
	}
}

func toByte(v uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	return buf[:]
}

/*
func TestHLLTC_Add_NoSparse(t *testing.T) {
	sk := NewTestSketch(16)
	sk.toNormal()

	sk.Insert(toByte(0x00010fffffffffff))
	n := sk.regs.regs[1]
	if n != 5 {
		t.Error(n)
	}

	sk.Insert(toByte(0x0002ffffffffffff))
	n = sk.regs.regs[2]
	if n != 1 {
		t.Error(n)
	}

	sk.Insert(toByte(0x0003000000000000))
	n = sk.regs.regs[3]
	if n != 49 {
		t.Error(n)
	}

	sk.Insert(toByte(0x0003000000000001))
	n = sk.regs.regs[3]
	if n != 49 {
		t.Error(n)
	}

	sk.Insert(toByte(0xff03700000000000))
	n = sk.regs.regs[0xff03]
	if n != 2 {
		t.Error(n)
	}

	sk.Insert(toByte(0xff03080000000000))
	n = sk.regs.regs[0xff03]
	if n != 5 {
		t.Error(n)
	}
}

func TestHLLTCPrecision_NoSparse(t *testing.T) {
	sk := NewTestSketch(4)
	sk.toNormal()

	sk.Insert(toByte(0x1fffffffffffffff))
	n := sk.regs.regs[1]
	if n != 1 {
		t.Error(n)
	}

	sk.Insert(toByte(0xffffffffffffffff))
	n = sk.regs.regs[0xf]
	if n != 1 {
		t.Error(n)
	}

	sk.Insert(toByte(0x00ffffffffffffff))
	n = sk.regs.regs[0]
	if n != 5 {
		t.Error(n)
	}
}

func TestHLLTC_toNormal(t *testing.T) {
	sk := NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.toNormal()
	c := sk.Estimate()
	if c != 1 {
		t.Error(c)
	}

	if sk.sparse {
		t.Error("toNormal should convert to normal")
	}

	sk = NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x0002ffffffffffff))
	sk.Insert(toByte(0x0003000000000000))
	sk.Insert(toByte(0x0003000000000001))
	sk.Insert(toByte(0xff03700000000000))
	sk.Insert(toByte(0xff03080000000000))
	sk.mergeSparse()
	sk.toNormal()

	n := sk.regs.regs[1]
	if n != 5 {
		t.Error(n)
	}
	n = sk.regs.regs[2]
	if n != 1 {
		t.Error(n)
	}
	n = sk.regs.regs[3]
	if n != 49 {
		t.Error(n)
	}
	n = sk.regs.regs[0xff03]
	if n != 5 {
		t.Error(n)
	}
}
*/
func TestHLLTCCount(t *testing.T) {
	sk := NewTestSketch(16)

	n := sk.Estimate()
	if n != 0 {
		t.Error(n)
	}

	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x00020fffffffffff))
	sk.Insert(toByte(0x00030fffffffffff))
	sk.Insert(toByte(0x00040fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))

	n = sk.Estimate()
	if n != 5 {
		t.Error(n)
	}

	// not mutated, still returns correct count
	n = sk.Estimate()
	if n != 5 {
		t.Error(n)
	}

	sk.Insert(toByte(0x00060fffffffffff))

	// mutated
	n = sk.Estimate()
	if n != 6 {
		t.Error(n)
	}
}

/*
func TestHLLTC_Merge_Error(t *testing.T) {
	sk := NewTestSketch(16)
	sk2 := NewTestSketch(10)

	err := sk.Merge(sk2)
	if err == nil {
		t.Error("different precision should return error")
	}
}

func TestHLL_Merge_Sparse(t *testing.T) {
	sk := NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x00020fffffffffff))
	sk.Insert(toByte(0x00030fffffffffff))
	sk.Insert(toByte(0x00040fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))

	sk2 := NewTestSketch(16)
	sk2.Merge(sk)
	n := sk2.Estimate()
	if n != 5 {
		t.Error(n)
	}

	if sk2.sparse {
		t.Error("Merge should convert to normal")
	}

	if !sk.sparse {
		t.Error("Merge should not modify argument")
	}

	sk2.Merge(sk)
	n = sk2.Estimate()
	if n != 5 {
		t.Error(n)
	}

	sk.Insert(toByte(0x00060fffffffffff))
	sk.Insert(toByte(0x00070fffffffffff))
	sk.Insert(toByte(0x00080fffffffffff))
	sk.Insert(toByte(0x00090fffffffffff))
	sk.Insert(toByte(0x000a0fffffffffff))
	sk.Insert(toByte(0x000a0fffffffffff))
	n = sk.Estimate()
	if n != 10 {
		t.Error(n)
	}

	sk2.Merge(sk)
	n = sk2.Estimate()
	if n != 10 {
		t.Error(n)
	}
}
*/

func TestHLLTC_EncodeDecode(t *testing.T) {
	sk := NewTestSketch(8)
	i, r := decodeHash(encodeHash(0xffffff8000000000, sk.p, pp), sk.p, pp)
	if i != 0xff {
		t.Error(i)
	}
	if r != 1 {
		t.Error(r)
	}

	i, r = decodeHash(encodeHash(0xff00000000000000, sk.p, pp), sk.p, pp)
	if i != 0xff {
		t.Error(i)
	}
	if r != 57 {
		t.Error(r)
	}

	i, r = decodeHash(encodeHash(0xff30000000000000, sk.p, pp), sk.p, pp)
	if i != 0xff {
		t.Error(i)
	}
	if r != 3 {
		t.Error(r)
	}

	i, r = decodeHash(encodeHash(0xaa10000000000000, sk.p, pp), sk.p, pp)
	if i != 0xaa {
		t.Error(i)
	}
	if r != 4 {
		t.Error(r)
	}

	i, r = decodeHash(encodeHash(0xaa0f000000000000, sk.p, pp), sk.p, pp)
	if i != 0xaa {
		t.Error(i)
	}
	if r != 5 {
		t.Error(r)
	}
}

func TestHLLTC_Error(t *testing.T) {
	_, err := New(3)
	if err == nil {
		t.Error("precision 3 should return error")
	}

	_, err = New(18)
	if err != nil {
		t.Error(err)
	}

	_, err = New(19)
	if err == nil {
		t.Error("precision 19 should return error")
	}
}

func TestHLLTC_Marshal_Unmarshal_Sparse(t *testing.T) {
	sk, _ := New(4)
	sk.sparse = true
	sk.tmpSet = map[uint32]struct{}{26: struct{}{}, 40: struct{}{}}

	// Add a bunch of values to the sparse representation.
	for i := 0; i < 10; i++ {
		sk.sparseList.Append(uint32(rand.Int()))
	}

	data, err := sk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Peeking at the first byte should reveal the version.
	if got, exp := data[0], byte(1); got != exp {
		t.Fatalf("got byte %v, expected %v", got, exp)
	}

	var res Sketch
	if err := res.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	// reflect.DeepEqual will always return false when comparing non-nil
	// functions, so we'll set them to nil.
	if got, exp := &res, sk; !reflect.DeepEqual(got, exp) {
		t.Fatalf("got %v, wanted %v", spew.Sdump(got), spew.Sdump(exp))
	}
}

func TestHLLTC_Marshal_Unmarshal_Dense(t *testing.T) {
	sk, _ := New(4)
	sk.sparse = false
	sk.toNormal()

	// Add a bunch of values to the dense representation.
	for i := uint32(0); i < 10; i++ {
		sk.regs.set(i, uint8(rand.Int()))
	}

	data, err := sk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Peeking at the first byte should reveal the version.
	if got, exp := data[0], byte(1); got != exp {
		t.Fatalf("got byte %v, expected %v", got, exp)
	}

	var res Sketch
	if err := res.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	// reflect.DeepEqual will always return false when comparing non-nil
	// functions, so we'll set them to nil.
	if got, exp := &res, sk; !reflect.DeepEqual(got, exp) {
		t.Fatalf("got %v, wanted %v", spew.Sdump(got), spew.Sdump(exp))
	}
}

// Tests that a sketch can be serialised / unserialised and keep an accurate
// cardinality estimate.
func TestHLLTC_Marshal_Unmarshal_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	count := make(map[string]struct{}, 1000000)
	sk, _ := New(16)

	buf := make([]byte, 8)
	for i := 0; i < 1000000; i++ {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}

		count[string(buf)] = struct{}{}

		// Add to the sketch.
		sk.Insert(buf)
	}

	gotC := sk.Estimate()
	epsilon := 15000 // 1.5%
	if got, exp := math.Abs(float64(int(gotC)-len(count))), epsilon; int(got) > exp {
		t.Fatalf("error was %v for estimation %d and true cardinality %d", got, gotC, len(count))
	}

	// Serialise the sketch.
	sketch, err := sk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Deserialise.
	sk = &Sketch{}
	if err := sk.UnmarshalBinary(sketch); err != nil {
		t.Fatal(err)
	}

	// The count should be the same
	oldC := gotC
	if got, exp := sk.Estimate(), oldC; got != exp {
		t.Fatalf("got %d, expected %d", got, exp)
	}

	// Add some more values.
	for i := 0; i < 1000000; i++ {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}

		count[string(buf)] = struct{}{}

		// Add to the sketch.
		sk.Insert(buf)
	}

	// The sketch should still be working correctly.
	gotC = sk.Estimate()
	epsilon = 30000 // 1.5%
	if got, exp := math.Abs(float64(int(gotC)-len(count))), epsilon; int(got) > exp {
		t.Fatalf("error was %v for estimation %d and true cardinality %d", got, gotC, len(count))
	}
}

func NewTestSketch(p uint8) *Sketch {
	sk, _ := New(p)
	return sk
}

// Generate random data to add to the sketch.
func genData(n int) [][]byte {
	out := make([][]byte, 0, n)
	buf := make([]byte, 8)

	for i := 0; i < n; i++ {
		// generate 8 random bytes
		n, err := rand.Read(buf)
		if err != nil {
			panic(err)
		} else if n != 8 {
			panic(fmt.Errorf("only %d bytes generated", n))
		}

		out = append(out, buf)
	}
	if len(out) != n {
		panic(fmt.Sprintf("wrong size slice: %d", n))
	}
	return out
}

// Memoises values to be added to a sketch during a benchmark.
var benchdata = map[int][][]byte{}

func benchmarkPlusAdd(b *testing.B, sk *Sketch, n int) {
	blobs, ok := benchdata[n]
	if !ok {
		// Generate it.
		benchdata[n] = genData(n)
		blobs = benchdata[n]
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(blobs); j++ {
			sk.Insert(blobs[j])
		}
	}
	b.StopTimer()
}

func BenchmarkPlus_Add_100(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 100)
}

func BenchmarkPlus_Add_1000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 1000)
}

func BenchmarkPlus_Add_10000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 10000)
}

func BenchmarkPlus_Add_100000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 100000)
}

func BenchmarkPlus_Add_1000000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 1000000)
}

func BenchmarkPlus_Add_10000000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 10000000)
}

func BenchmarkPlus_Add_100000000(b *testing.B) {
	sk, _ := New(16)
	benchmarkPlusAdd(b, sk, 100000000)
}
