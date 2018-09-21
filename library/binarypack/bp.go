/*
	Package binary_pack performs conversions between some Go values represented as byte slices.
	This can be used in handling binary data stored in files or from network connections,
	among other sources. It uses format slices of strings as compact descriptions of the layout
	of the Go structs.
	Format characters (some characters like H have been reserved for future implementation of unsigned numbers):
		? - bool, packed size 1 byte
		h, H - int, packed size 2 bytes (in future it will support binarypack/unpack of int8, uint8 values)
		i, I, l, L - int, packed size 4 bytes (in future it will support binarypack/unpack of int16, uint16, int32, uint32 values)
		q, Q - int, packed size 8 bytes (in future it will support binarypack/unpack of int64, uint64 values)
		f - float32, packed size 4 bytes
		d - float64, packed size 8 bytes
		Ns - string, packed size N bytes, N is a number of runes to binarypack/unpack
*/

package binarypack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type BinaryPack struct{}

// Return a byte slice containing the values of msg slice packed according to the given format.
// The items of msg slice must match the values required by the format exactly.
func (bp *BinaryPack) Pack(format []string, msg []interface{}) (res []byte, err error) {
	var (
		order binary.ByteOrder
		i     int
	)

	if formatLen(format) > len(msg) {
		err = errors.New("Format is longer than values to binarypack")
		return
	}

	//switch t := msg[1].(type) {
	//default:
	//	log.Error("type is: %T\n", t)
	//}
	order = binary.LittleEndian
	for _, f := range format {
		switch f {
		case "<":
			order = binary.LittleEndian
			i -= 1
		case ">", "!":
			order = binary.BigEndian
			i -= 1
		case "?":
			casted_value, ok := msg[i].(bool)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (bool)")
				return
			}
			res = append(res, boolToBytes(casted_value, order)...)
		case "h", "H":
			casted_value, ok := msg[i].(int64)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (int64, 2 bytes)")
				return
			}
			res = append(res, int64ToBytes(casted_value, 2, order)...)
		case "i", "I", "l", "L":
			casted_value, ok := msg[i].(int64)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (int64, 4 bytes)")
				return
			}
			res = append(res, int64ToBytes(casted_value, 4, order)...)
		case "q", "Q":
			casted_value, ok := msg[i].(int64)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (int64, 8 bytes)")
				return
			}
			res = append(res, int64ToBytes(casted_value, 8, order)...)
		case "f":
			casted_value, ok := msg[i].(float32)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (float32)")
				return
			}
			res = append(res, float32ToBytes(casted_value, 4, order)...)
		case "d":
			casted_value, ok := msg[i].(float64)
			if !ok {
				err = errors.New("Type of passed value doesn't match to expected '" + f + "' (float64)")
				return
			}
			res = append(res, float64ToBytes(casted_value, 8, order)...)
		default:
			if strings.Contains(f, "s") {
				casted_value, ok := msg[i].(string)
				if !ok {
					err = errors.New("Type of passed value doesn't match to expected '" + f + "' (string)")
					return
				}
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				l := len(casted_value)
				size := 0
				if n > l {
					size = n - l
				}
				if order == binary.BigEndian {
					res = append(res, []byte(fmt.Sprintf("%s%s",
						strings.Repeat("\x00", size), reverse(casted_value[:n-size])))...)
				} else {
					res = append(res, []byte(fmt.Sprintf("%s%s",
						casted_value[:n-size], strings.Repeat("\x00", size)))...)
				}
			} else {
				err = errors.New("Unexpected format token: '" + f + "'")
				return
			}
		}
		i++
	}

	return
}

// Unpack the byte slice (presumably packed by Pack(format, msg)) according to the given format.
// The result is a []interface{} slice even if it contains exactly one item.
// The byte slice must contain not less the amount of data required by the format
// (len(msg) must more or equal CalcSize(format)).
func (bp *BinaryPack) UnPack(format []string, msg []byte) (res []interface{}, err error) {
	var (
		order         binary.ByteOrder
		expected_size int
	)

	if expected_size, err = bp.CalcSize(format); err != nil {
		return
	}

	if expected_size > len(msg) {
		err = errors.New("Expected size is bigger than actual size of message")
		return
	}

	order = binary.LittleEndian
	for _, f := range format {
		switch f {
		case "<":
			order = binary.LittleEndian
		case ">", "!":
			order = binary.BigEndian
		case "?":
			res = append(res, bytesToBool(msg[:1], order))
			msg = msg[1:]
		case "h", "H":
			res = append(res, bytesToInt64(msg[:2], order))
			msg = msg[2:]
		case "i", "I", "l", "L":
			res = append(res, bytesToInt64(msg[:4], order))
			msg = msg[4:]
		case "q", "Q":
			res = append(res, bytesToInt64(msg[:8], order))
			msg = msg[8:]
		case "f":
			res = append(res, bytesToFloat32(msg[:4], order))
			msg = msg[4:]
		case "d":
			res = append(res, bytesToFloat64(msg[:8], order))
			msg = msg[8:]
		default:
			if strings.Contains(f, "s") {
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				if order == binary.BigEndian {
					res = append(res, strings.TrimRight(reverse(string(msg[:n])), "\x00"))
				} else {
					res = append(res, strings.TrimRight(string(msg[:n]), "\x00"))
				}
				msg = msg[n:]
			} else {
				return nil, errors.New("Unexpected format token: '" + f + "'")
			}
		}
	}

	return res, nil
}

// Return the size of the struct (and hence of the byte slice) corresponding to the given format.
func (bp *BinaryPack) CalcSize(format []string) (int, error) {
	var size int

	for _, f := range format {
		switch f {
		case "<", ">", "!":
			size += 0
		case "?":
			size = size + 1
		case "h", "H":
			size = size + 2
		case "i", "I", "l", "L", "f":
			size = size + 4
		case "q", "Q", "d":
			size = size + 8
		default:
			if strings.Contains(f, "s") {
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				size = size + n
			} else {
				return 0, errors.New("Unexpected format token: '" + f + "'")
			}
		}
	}

	return size, nil
}

func formatLen(format []string) (length int) {
	for _, f := range format {
		switch f {
		case "@", "=", "<", ">", "!":
			length += 0
		default:
			length++
		}
	}

	return
}

func boolToBytes(x bool, order binary.ByteOrder) []byte {
	if x {
		return int64ToBytes(1, 1, order)
	}
	return int64ToBytes(0, 1, order)
}

func bytesToBool(b []byte, order binary.ByteOrder) bool {
	return bytesToInt64(b, order) > 0
}

func int64ToBytes(n int64, size int, order binary.ByteOrder) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, order, int64(n))
	if order == binary.BigEndian {
		return buf.Bytes()[8-size:]
	}
	return buf.Bytes()[0:size]
}

func bytesToInt64(b []byte, order binary.ByteOrder) int64 {
	buf := bytes.NewBuffer(b)

	switch len(b) {
	case 1:
		var x int8
		binary.Read(buf, order, &x)
		return int64(x)
	case 2:
		var x int16
		binary.Read(buf, order, &x)
		return int64(x)
	case 4:
		var x int32
		binary.Read(buf, order, &x)
		return int64(x)
	default:
		var x int64
		binary.Read(buf, order, &x)
		return int64(x)
	}
}

func float32ToBytes(n float32, size int, order binary.ByteOrder) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, order, n)
	if order == binary.BigEndian {
		return buf.Bytes()[4-size:]
	}
	return buf.Bytes()[0:size]
}

func bytesToFloat32(b []byte, order binary.ByteOrder) float32 {
	var x float32
	buf := bytes.NewBuffer(b)
	binary.Read(buf, order, &x)
	return x
}

func float64ToBytes(n float64, size int, order binary.ByteOrder) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, order, n)
	if order == binary.BigEndian {
		return buf.Bytes()[8-size:]
	}
	return buf.Bytes()[0:size]
}

func bytesToFloat64(b []byte, order binary.ByteOrder) float64 {
	var x float64
	buf := bytes.NewBuffer(b)
	binary.Read(buf, order, &x)
	return x
}

func reverse(s string) string {
	cs := make([]rune, utf8.RuneCountInString(s))
	i := len(cs)
	for _, c := range s {
		i--
		cs[i] = c
	}
	return string(cs)
}
