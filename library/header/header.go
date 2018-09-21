package header

import "github.com/eyotang/load/library/binarypack"

type Header struct {
	bp *binarypack.BinaryPack
}

func New() *Header {
	return &Header{
		bp: new(binarypack.BinaryPack),
	}
}

func (h *Header) Pack(format []string, headers []interface{}) (header []byte, err error) {
	if header, err = h.bp.Pack(format, headers); err != nil {
		return
	}
	return
}

func (h *Header) Unpack(format []string, header []byte) (res []interface{}, err error) {
	if res, err = h.bp.UnPack(format, header); err != nil {
		return
	}
	return
}
