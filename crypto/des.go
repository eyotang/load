package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"

	"github.com/pkg/errors"
)

const (
	// Modes of crypting / cyphering
	ECB = 0
	CBC = 1

	// Modes of padding
	PAD_NORMAL = 1
	PAD_PKCS5  = 2
)

// Des def
type Des struct {
	mode    uint8
	padmode uint8
	b       cipher.Block
	ebm     cipher.BlockMode
	dbm     cipher.BlockMode
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(src []byte) []byte {
	return bytes.TrimFunc(src,
		func(r rune) bool {
			return r == rune(0)
		})
}

func NewDes(key []byte, mode uint8, iv []byte, padmode uint8) (d *Des, err error) {
	var (
		b   cipher.Block
		ebm cipher.BlockMode
		dbm cipher.BlockMode
	)

	if b, err = des.NewCipher(key); err != nil {
		err = errors.Wrapf(err, "des.NewCipher failed")
		return
	}

	if mode == CBC {
		ebm = cipher.NewCBCEncrypter(b, iv)
		dbm = cipher.NewCBCDecrypter(b, iv)
	}

	d = &Des{
		b:       b,
		ebm:     ebm,
		dbm:     dbm,
		mode:    mode,
		padmode: padmode,
	}

	return
}

func (d *Des) Encrypt(plaintext []byte) (ciphertext []byte) {
	var (
		bs  = d.b.BlockSize()
		dst []byte
	)

	if d.padmode == PAD_PKCS5 {
		plaintext = PKCS5Padding(plaintext, bs)
	} else {
		plaintext = ZeroPadding(plaintext, bs)
	}

	ciphertext = make([]byte, len(plaintext))

	if d.mode == CBC {
		d.ebm.CryptBlocks(ciphertext, plaintext)
	} else if d.mode == ECB {
		dst = ciphertext
		for len(plaintext) > 0 {
			d.b.Encrypt(dst, plaintext[:bs])
			plaintext = plaintext[bs:]
			dst = dst[bs:]
		}
	}

	return
}

func (d *Des) Decrypt(ciphertext []byte) (plaintext []byte) {
	var (
		bs  = d.b.BlockSize()
		dst []byte
	)

	plaintext = make([]byte, len(ciphertext))

	if d.mode == CBC {
		d.dbm.CryptBlocks(plaintext, ciphertext)
	} else if d.mode == ECB {
		dst = plaintext
		for len(ciphertext) > 0 {
			d.b.Decrypt(dst, ciphertext[:bs])
			ciphertext = ciphertext[bs:]
			dst = dst[bs:]
		}
	}

	if d.padmode == PAD_PKCS5 {
		plaintext = PKCS5UnPadding(plaintext)
	} else {
		plaintext = ZeroUnPadding(plaintext)
	}

	return
}
