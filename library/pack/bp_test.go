package tcp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBinaryPack_CalcSize(t *testing.T) {
	type Case struct {
		in   []string
		want int
		e    bool
	}
	cases := []Case{
		{[]string{}, 0, false},
		{[]string{"I", "I", "I", "4s"}, 16, false},
		{[]string{"H", "H", "I", "H", "8s", "H"}, 20, false},
		{[]string{"i", "?", "H", "f", "d", "h", "I", "5s"}, 30, false},
		{[]string{"?", "h", "H", "i", "I", "l", "L", "q", "Q", "f", "d", "1s"}, 50, false},
	}
	invalids := []Case{
		// Unknown tokens
		{[]string{"a", "b", "c"}, 0, true},
	}

	Convey("TEST CalcSize", t, func() {
		for _, c := range cases {
			got, err := new(BinaryPack).CalcSize(c.in)
			So(err, ShouldBeNil)
			So(got, ShouldEqual, c.want)
		}

		for _, c := range invalids {
			got, err := new(BinaryPack).CalcSize(c.in)
			So(err, ShouldNotBeNil)
			So(got, ShouldEqual, c.want)
		}
	})
}

func TestBinaryPack_Pack(t *testing.T) {
	type Case struct {
		f    []string
		a    []interface{}
		want []byte
	}
	cases := []Case{
		{[]string{"?", "?"}, []interface{}{true, false}, []byte{1, 0}},
		{[]string{">", "?", "?"}, []interface{}{true, false}, []byte{1, 0}},
		{[]string{"h", "h", "h"}, []interface{}{int64(0), int64(5), int64(-5)},
			[]byte{0, 0, 5, 0, 251, 255}},
		{[]string{"!", "h", "h", "h"}, []interface{}{int64(0), int64(5), int64(-5)},
			[]byte{0, 0, 0, 5, 255, 251}},
		{[]string{"H", "H", "H"}, []interface{}{int64(0), int64(5), int64(2300)},
			[]byte{0, 0, 5, 0, 252, 8}},
		{[]string{"<", "H", "H", "H"}, []interface{}{int64(0), int64(5), int64(2300)},
			[]byte{0, 0, 5, 0, 252, 8}},
		{[]string{">", "H", "H", "H"}, []interface{}{int64(0), int64(5), int64(2300)},
			[]byte{0, 0, 0, 5, 8, 252}},
		{[]string{"i", "i", "i"}, []interface{}{int64(0), int64(5), int64(-5)},
			[]byte{0, 0, 0, 0, 5, 0, 0, 0, 251, 255, 255, 255}},
		{[]string{">", "i", "i", "i"}, []interface{}{int64(0), int64(5), int64(-5)},
			[]byte{0, 0, 0, 0, 0, 0, 0, 5, 255, 255, 255, 251}},
		{[]string{"I", "I", "I"}, []interface{}{int64(0), int64(5), int64(2300)},
			[]byte{0, 0, 0, 0, 5, 0, 0, 0, 252, 8, 0, 0}},
		{[]string{"f", "f", "f"}, []interface{}{float32(0.0), float32(5.3), float32(-5.3)},
			[]byte{0, 0, 0, 0, 154, 153, 169, 64, 154, 153, 169, 192}},
		{[]string{">", "f", "f", "f"}, []interface{}{float32(0.0), float32(5.3), float32(-5.3)},
			[]byte{0, 0, 0, 0, 64, 169, 153, 154, 192, 169, 153, 154}},
		{[]string{"d", "d", "d"}, []interface{}{0.0, 5.3, -5.3},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 51, 51, 51, 51, 51, 51, 21, 64, 51, 51, 51, 51, 51, 51, 21, 192}},
		{[]string{"!", "d", "d", "d"}, []interface{}{0.0, 5.3, -5.3},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 64, 21, 51, 51, 51, 51, 51, 51, 192, 21, 51, 51, 51, 51, 51, 51}},
		{[]string{"1s", "2s", "10s"}, []interface{}{"a", "be", "1234567890"},
			[]byte{97, 98, 101, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48}},
		{[]string{">", "1s", "2s", "10s"}, []interface{}{"a", "be", "1234567890"},
			[]byte{97, 101, 98, 48, 57, 56, 55, 54, 53, 52, 51, 50, 49}},
		{[]string{"I", "I", "I", "4s"}, []interface{}{int64(1), int64(2), int64(4), "DUMP"},
			[]byte{1, 0, 0, 0, 2, 0, 0, 0, 4, 0, 0, 0, 68, 85, 77, 80}},
		{[]string{"!", "I", "I", "I", "4s"}, []interface{}{int64(1), int64(2), int64(4), "DUMP"},
			[]byte{0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 4, 80, 77, 85, 68}},
		{[]string{"i", "h", "d", "5s"}, []interface{}{int64(1), int64(2), 4.8, "DUMP"},
			[]byte{1, 0, 0, 0, 2, 0, 51, 51, 51, 51, 51, 51, 19, 64, 68, 85, 77, 80, 0}},
		{[]string{"!", "i", "h", "d", "5s"}, []interface{}{int64(1), int64(2), 4.8, "DUMP"},
			[]byte{0, 0, 0, 1, 0, 2, 64, 19, 51, 51, 51, 51, 51, 51, 0, 80, 77, 85, 68}},
		{[]string{"i", "h", "d", "3s"}, []interface{}{int64(1), int64(2), 453.8, "DUMP"},
			[]byte{1, 0, 0, 0, 2, 0, 205, 204, 204, 204, 204, 92, 124, 64, 68, 85, 77}},
		{[]string{"!", "i", "h", "d", "3s"}, []interface{}{int64(1), int64(2), 453.8, "DUMP"},
			[]byte{0, 0, 0, 1, 0, 2, 64, 124, 92, 204, 204, 204, 204, 205, 77, 85, 68}},
	}

	Convey("TEST Pack", t, func() {
		for _, c := range cases {
			got, err := new(BinaryPack).Pack(c.f, c.a)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, c.want)
		}
	})

	invalids := []Case{
		// Wrong format length
		{[]string{"I", "I", "I", "4s"}, []interface{}{1, 4, "DUMP"}, nil},
		// Wrong format token
		{[]string{"I", "a", "I", "4s"}, []interface{}{1, 2, 4, "DUMP"}, nil},
		// Wrong types
		{[]string{"?"}, []interface{}{1.0}, nil},
		{[]string{"H"}, []interface{}{int8(1)}, nil},
		{[]string{"I"}, []interface{}{int32(2)}, nil},
		{[]string{"Q"}, []interface{}{int(3)}, nil},
		{[]string{"f"}, []interface{}{float64(2.5)}, nil},
		{[]string{"d"}, []interface{}{float32(2.5)}, nil},
		{[]string{"1s"}, []interface{}{'a'}, nil},
	}

	Convey("TEST Pack invalid", t, func() {
		for _, c := range invalids {
			_, err := new(BinaryPack).Pack(c.f, c.a)
			So(err, ShouldNotBeNil)
		}
	})
}

func TestBinaryPack_UnPack(t *testing.T) {
	type Case struct {
		f    []string
		a    []byte
		want []interface{}
	}
	cases := []Case{
		{[]string{"?", "?"}, []byte{1, 0}, []interface{}{true, false}},
		{[]string{">", "?", "?"}, []byte{1, 0}, []interface{}{true, false}},
		{[]string{"h", "h", "h"}, []byte{0, 0, 5, 0, 251, 255},
			[]interface{}{int64(0), int64(5), int64(-5)}},
		{[]string{"!", "h", "h", "h"}, []byte{0, 0, 0, 5, 255, 251},
			[]interface{}{int64(0), int64(5), int64(-5)}},
		{[]string{"H", "H", "H"}, []byte{0, 0, 5, 0, 252, 8},
			[]interface{}{int64(0), int64(5), int64(2300)}},
		{[]string{">", "H", "H", "H"}, []byte{0, 0, 0, 5, 8, 252},
			[]interface{}{int64(0), int64(5), int64(2300)}},
		{[]string{"i", "i", "i"}, []byte{0, 0, 0, 0, 5, 0, 0, 0, 251, 255, 255, 255},
			[]interface{}{int64(0), int64(5), int64(-5)}},
		{[]string{"!", "i", "i", "i"}, []byte{0, 0, 0, 0, 0, 0, 0, 5, 255, 255, 255, 251},
			[]interface{}{int64(0), int64(5), int64(-5)}},
		{[]string{"I", "I", "I"}, []byte{0, 0, 0, 0, 5, 0, 0, 0, 252, 8, 0, 0},
			[]interface{}{int64(0), int64(5), int64(2300)}},
		{[]string{">", "I", "I", "I"}, []byte{0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 8, 252},
			[]interface{}{int64(0), int64(5), int64(2300)}},
		{[]string{"f", "f", "f"},
			[]byte{0, 0, 0, 0, 154, 153, 169, 64, 154, 153, 169, 192},
			[]interface{}{float32(0.0), float32(5.3), float32(-5.3)}},
		{[]string{"!", "f", "f", "f"},
			[]byte{0, 0, 0, 0, 64, 169, 153, 154, 192, 169, 153, 154},
			[]interface{}{float32(0.0), float32(5.3), float32(-5.3)}},
		{[]string{"d", "d", "d"},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 51, 51, 51, 51, 51, 51, 21, 64, 51, 51, 51, 51, 51, 51, 21, 192},
			[]interface{}{0.0, 5.3, -5.3}},
		{[]string{">", "d", "d", "d"},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 64, 21, 51, 51, 51, 51, 51, 51, 192, 21, 51, 51, 51, 51, 51, 51},
			[]interface{}{0.0, 5.3, -5.3}},
		{[]string{"1s", "2s", "10s"},
			[]byte{97, 98, 101, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
			[]interface{}{"a", "be", "1234567890"}},
		{[]string{"!", "1s", "2s", "10s"},
			[]byte{97, 101, 98, 48, 57, 56, 55, 54, 53, 52, 51, 50, 49},
			[]interface{}{"a", "be", "1234567890"}},
		{[]string{"I", "I", "I", "4s"},
			[]byte{1, 0, 0, 0, 2, 0, 0, 0, 4, 0, 0, 0, 68, 85, 77, 80},
			[]interface{}{int64(1), int64(2), int64(4), "DUMP"}},
		{[]string{">", "I", "I", "I", "4s"},
			[]byte{0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 4, 80, 77, 85, 68},
			[]interface{}{int64(1), int64(2), int64(4), "DUMP"}},
		{[]string{"i", "h", "d", "5s"},
			[]byte{1, 0, 0, 0, 2, 0, 51, 51, 51, 51, 51, 51, 19, 64, 68, 85, 77, 80, 0},
			[]interface{}{int64(1), int64(2), 4.8, "DUMP"}},
		{[]string{"!", "i", "h", "d", "5s"},
			[]byte{0, 0, 0, 1, 0, 2, 64, 19, 51, 51, 51, 51, 51, 51, 0, 80, 77, 85, 68},
			[]interface{}{int64(1), int64(2), 4.8, "DUMP"}},
		{[]string{"i", "h", "d", "3s"},
			[]byte{1, 0, 0, 0, 2, 0, 205, 204, 204, 204, 204, 92, 124, 64, 68, 85, 77, 0},
			[]interface{}{int64(1), int64(2), 453.8, "DUM"}},
		{[]string{">", "i", "h", "d", "3s"},
			[]byte{0, 0, 0, 1, 0, 2, 64, 124, 92, 204, 204, 204, 204, 205, 77, 85, 68, 0},
			[]interface{}{int64(1), int64(2), 453.8, "DUM"}},
	}

	Convey("TEST UnPack", t, func() {
		for _, c := range cases {
			got, err := new(BinaryPack).UnPack(c.f, c.a)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, c.want)
		}
	})

	invalids := []Case{
		// Wrong format length
		{[]string{"I", "I", "I", "4s", "H"}, []byte{1, 0, 0, 0, 2, 0, 0, 0, 4, 0, 0, 0, 68, 85, 77, 80},
			nil},
		// Wrong format token
		{[]string{"I", "a", "I", "4s"}, []byte{1, 0, 0, 0, 2, 0, 0, 0, 4, 0, 0, 0, 68, 85, 77, 80},
			nil},
	}

	Convey("TEST UnPack invalid", t, func() {
		for _, c := range invalids {
			_, err := new(BinaryPack).UnPack(c.f, c.a)
			So(err, ShouldNotBeNil)
		}
	})
}
