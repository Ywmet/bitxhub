package blockfile

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/stretchr/testify/assert"
)

func getChunk(size int, b int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(b)
	}
	return data
}

func TestBlockFileBasics(t *testing.T) {
	// set cutoff at 50 bytes
	f, err := newTable(os.TempDir(),
		fmt.Sprintf("unittest-%d", rand.Uint64()), 2*1000*1000*1000, log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()
	// Write 15 bytes 255 times, results in 85 files
	for x := 0; x < 255; x++ {
		data := getChunk(15, x)
		f.Append(uint64(x), data)
	}
	for y := 0; y < 255; y++ {
		exp := getChunk(15, y)
		got, err := f.Retrieve(uint64(y))
		assert.Nil(t, err)
		if !bytes.Equal(got, exp) {
			t.Fatalf("test %d, got \n%x != \n%x", y, got, exp)
		}
	}
	// Check that we cannot read too far
	_, err = f.Retrieve(uint64(255))
	assert.Equal(t, fmt.Errorf("out of bounds"), err)
}

func TestFreezerBasicsClosing(t *testing.T) {
	var (
		fname  = fmt.Sprintf("basics-close-%d", rand.Uint64())
		logger = log.NewWithModule("blockfile_test")
		f      *BlockTable
		err    error
	)
	f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
	assert.Nil(t, err)
	// Write 15 bytes 255 times, results in 85 files
	for x := 0; x < 255; x++ {
		data := getChunk(15, x)
		f.Append(uint64(x), data)
		f.Close()
		f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
		assert.Nil(t, err)
	}
	defer f.Close()

	for y := 0; y < 255; y++ {
		exp := getChunk(15, y)
		got, err := f.Retrieve(uint64(y))
		assert.Nil(t, err)
		if !bytes.Equal(got, exp) {
			t.Fatalf("test %d, got \n%x != \n%x", y, got, exp)
		}
		f.Close()
		f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
		assert.Nil(t, err)
	}
}

func TestFreezerTruncate(t *testing.T) {
	fname := fmt.Sprintf("truncation-%d", rand.Uint64())
	logger := log.NewWithModule("blockfile_test")

	{ // Fill table
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 0; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.items - 1)
		assert.Nil(t, err)
		f.Close()
	}
	// Reopen, truncate
	{
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		defer f.Close()
		// for x := 0; x < 20; x++ {
		// 	f.truncate(uint64(30 - x - 1)) // 150 bytes
		// }
		f.truncate(10)
		if f.items != 10 {
			t.Fatalf("expected %d items, got %d", 10, f.items)
		}
		// 45, 45, 45, 15 -- bytes should be 15
		if f.headBytes != 15 {
			t.Fatalf("expected %d bytes, got %d", 15, f.headBytes)
		}

	}
}

func TestFreezerReadAndTruncate(t *testing.T) {
	fname := fmt.Sprintf("read_truncate-%d", rand.Uint64())
	logger := log.NewWithModule("blockfile_test")
	{ // Fill table
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 0; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.items - 1)
		assert.Nil(t, err)
		f.Close()
	}
	// Reopen and read all files
	{
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		if f.items != 30 {
			f.Close()
			t.Fatalf("expected %d items, got %d", 0, f.items)
		}
		for y := byte(0); y < 30; y++ {
			f.Retrieve(uint64(y))
		}
		// Now, truncate back to zero
		f.truncate(0)
		// Write the data again
		for x := 0; x < 30; x++ {
			data := getChunk(15, ^x)
			err := f.Append(uint64(x), data)
			assert.Nil(t, err)
		}
		f.Close()
	}
}
