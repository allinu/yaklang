// Copyright 2015, 2018, 2019 Opsmate, Inc. All rights reserved.
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkcs12

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"hash"
	"io"

	"github.com/yaklang/yaklang/common/utils/tlsutils/go-pkcs12/rc2"

	"golang.org/x/crypto/pbkdf2"
)

var (
	oidPBEWithSHAAnd3KeyTripleDESCBC = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 3})
	oidPBEWithSHAAnd128BitRC4        = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 1})
	oidPBEWithSHAAnd128BitRC2CBC     = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 5})
	oidPBEWithSHAAnd40BitRC2CBC      = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 6})
	oidPBEWithSHAAndDES              = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 5, 10})
	oidPBES2                         = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 5, 13})
	oidPBKDF2                        = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 5, 12})
	oidHmacWithSHA1                  = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 2, 7})
	oidHmacWithSHA256                = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 2, 9})
	oidHmacWithSHA512                = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 2, 11})
	oidHmacWithMD5                   = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 2, 5})
	oidAES128CBC                     = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 1, 2})
	oidAES192CBC                     = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 1, 22})
	oidAES256CBC                     = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 1, 42})
)

// pbeCipher is an abstraction of a PKCS#12 cipher.
type pbeCipher interface {
	// create returns a cipher.Block given a key.
	create(key []byte) (cipher.Block, error)
	// deriveKey returns a key derived from the given password and salt.
	deriveKey(salt, password []byte, iterations int) []byte
	// deriveKey returns an IV derived from the given password and salt.
	deriveIV(salt, password []byte, iterations int) []byte
}

type shaWithTripleDESCBC struct{}

func (shaWithTripleDESCBC) create(key []byte) (cipher.Block, error) {
	return des.NewTripleDESCipher(key)
}

func (shaWithTripleDESCBC) deriveKey(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 1, 24)
}

func (shaWithTripleDESCBC) deriveIV(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 2, 8)
}

type shaWith128BitRC2CBC struct{}

func (shaWith128BitRC2CBC) create(key []byte) (cipher.Block, error) {
	return rc2.New(key, len(key)*8)
}

func (shaWith128BitRC2CBC) deriveKey(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 1, 16)
}

func (shaWith128BitRC2CBC) deriveIV(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 2, 8)
}

type shaWith40BitRC2CBC struct{}

func (shaWith40BitRC2CBC) create(key []byte) (cipher.Block, error) {
	return rc2.New(key, len(key)*8)
}

func (shaWith40BitRC2CBC) deriveKey(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 1, 5)
}

func (shaWith40BitRC2CBC) deriveIV(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 2, 8)
}

type shaWith128BitRC4 struct{}

func (shaWith128BitRC4) create(key []byte) (cipher.Block, error) {
	// 由于RC4是流密码而非块密码，我们返回一个特殊标记
	return nil, nil
}

func (shaWith128BitRC4) deriveKey(salt, password []byte, iterations int) []byte {
	return pbkdf(sha1Sum, 20, 64, salt, password, iterations, 1, 16)
}

func (shaWith128BitRC4) deriveIV(salt, password []byte, iterations int) []byte {
	// RC4不使用IV
	return nil
}

// rc4CipherImpl结构不再需要
// type rc4CipherImpl struct {}

type pbeParams struct {
	Salt       []byte
	Iterations int
}

func pbeCipherFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.Block, []byte, error) {
	var cipherType pbeCipher

	switch {
	case algorithm.Algorithm.Equal(oidPBEWithSHAAnd3KeyTripleDESCBC):
		cipherType = shaWithTripleDESCBC{}
	case algorithm.Algorithm.Equal(oidPBEWithSHAAnd128BitRC2CBC):
		cipherType = shaWith128BitRC2CBC{}
	case algorithm.Algorithm.Equal(oidPBEWithSHAAnd40BitRC2CBC):
		cipherType = shaWith40BitRC2CBC{}
	case algorithm.Algorithm.Equal(oidPBEWithSHAAnd128BitRC4):
		cipherType = shaWith128BitRC4{}
	case algorithm.Algorithm.Equal(oidPBEWithSHAAndDES):
		// PKCS#5 PBE-SHA1-DES uses PBKDF1, not PKCS#12 PBKDF
		return pbkdf1CipherFor(algorithm, password)
	case algorithm.Algorithm.Equal(oidPBES2):
		// rfc7292#appendix-B.1 (the original PKCS#12 PBE) requires passwords formatted as BMPStrings.
		// However, rfc8018#section-3 recommends that the password for PBES2 follow ASCII or UTF-8.
		// This is also what Windows expects.
		// Therefore, we convert the password to UTF-8.
		originalPassword, err := decodeBMPString(password)
		if err != nil {
			return nil, nil, err
		}
		utf8Password := []byte(originalPassword)
		return pbes2CipherFor(algorithm, utf8Password)
	default:
		return nil, nil, NotImplementedError("pbe algorithm " + algorithm.Algorithm.String() + " is not supported")
	}

	var params pbeParams
	if err := unmarshal(algorithm.Parameters.FullBytes, &params); err != nil {
		return nil, nil, err
	}

	key := cipherType.deriveKey(params.Salt, password, params.Iterations)
	iv := cipherType.deriveIV(params.Salt, password, params.Iterations)

	// 检查是否是RC4算法
	if algorithm.Algorithm.Equal(oidPBEWithSHAAnd128BitRC4) {
		// RC4特殊处理，返回nil作为block，key作为数据
		return nil, key, nil
	}

	block, err := cipherType.create(key)
	if err != nil {
		return nil, nil, err
	}

	return block, iv, nil
}

func pbDecrypterFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.BlockMode, int, error) {
	block, iv, err := pbeCipherFor(algorithm, password)
	if err != nil {
		return nil, 0, err
	}

	// RC4特殊处理
	if algorithm.Algorithm.Equal(oidPBEWithSHAAnd128BitRC4) {
		// RC4不使用CBC模式，我们返回nil和特殊标记
		return nil, -1, nil
	}

	return cipher.NewCBCDecrypter(block, iv), block.BlockSize(), nil
}

func pbDecrypt(info decryptable, password []byte) (decrypted []byte, err error) {
	// 检查是否是RC4算法
	if info.Algorithm().Algorithm.Equal(oidPBEWithSHAAnd128BitRC4) {
		// RC4特殊处理
		_, key, err := pbeCipherFor(info.Algorithm(), password)
		if err != nil {
			return nil, err
		}

		// 使用RC4密钥进行解密
		encrypted := info.Data()
		if len(encrypted) == 0 {
			return nil, errors.New("pkcs12: empty encrypted data")
		}

		// 创建真正的RC4密码流
		rc4Cipher, err := rc4.NewCipher(key)
		if err != nil {
			return nil, err
		}

		// RC4解密
		decrypted = make([]byte, len(encrypted))
		rc4Cipher.XORKeyStream(decrypted, encrypted)
		return decrypted, nil
	}

	cbc, blockSize, err := pbDecrypterFor(info.Algorithm(), password)
	if err != nil {
		return nil, err
	}

	encrypted := info.Data()
	if len(encrypted) == 0 {
		return nil, errors.New("pkcs12: empty encrypted data")
	}
	if len(encrypted)%blockSize != 0 {
		return nil, errors.New("pkcs12: input is not a multiple of the block size")
	}
	decrypted = make([]byte, len(encrypted))
	cbc.CryptBlocks(decrypted, encrypted)

	psLen := int(decrypted[len(decrypted)-1])
	if psLen == 0 || psLen > blockSize {
		return nil, ErrDecryption
	}

	if len(decrypted) < psLen {
		return nil, ErrDecryption
	}

	ps := decrypted[len(decrypted)-psLen:]
	decrypted = decrypted[:len(decrypted)-psLen]
	if bytes.Compare(ps, bytes.Repeat([]byte{byte(psLen)}, psLen)) != 0 {
		return nil, ErrDecryption
	}

	return
}

//	PBES2-params ::= SEQUENCE {
//		keyDerivationFunc AlgorithmIdentifier {{PBES2-KDFs}},
//		encryptionScheme AlgorithmIdentifier {{PBES2-Encs}}
//	}
type pbes2Params struct {
	Kdf              pkix.AlgorithmIdentifier
	EncryptionScheme pkix.AlgorithmIdentifier
}

//	PBKDF2-params ::= SEQUENCE {
//	    salt CHOICE {
//	      specified OCTET STRING,
//	      otherSource AlgorithmIdentifier {{PBKDF2-SaltSources}}
//	    },
//	    iterationCount INTEGER (1..MAX),
//	    keyLength INTEGER (1..MAX) OPTIONAL,
//	    prf AlgorithmIdentifier {{PBKDF2-PRFs}} DEFAULT
//	    algid-hmacWithSHA1
//	}
type pbkdf2Params struct {
	Salt       asn1.RawValue
	Iterations int
	KeyLength  int                      `asn1:"optional"`
	Prf        pkix.AlgorithmIdentifier `asn1:"optional"`
}

func pbes2CipherFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.Block, []byte, error) {
	var params pbes2Params
	if err := unmarshal(algorithm.Parameters.FullBytes, &params); err != nil {
		return nil, nil, err
	}

	if !params.Kdf.Algorithm.Equal(oidPBKDF2) {
		return nil, nil, NotImplementedError("pbes2 kdf algorithm " + params.Kdf.Algorithm.String() + " is not supported")
	}

	var kdfParams pbkdf2Params
	if err := unmarshal(params.Kdf.Parameters.FullBytes, &kdfParams); err != nil {
		return nil, nil, err
	}
	if kdfParams.Salt.Tag != asn1.TagOctetString {
		return nil, nil, NotImplementedError("only octet string salts are supported for pbes2/pbkdf2")
	}

	var prf func() hash.Hash
	switch {
	case kdfParams.Prf.Algorithm.Equal(oidHmacWithSHA256):
		prf = sha256.New
	case kdfParams.Prf.Algorithm.Equal(oidHmacWithSHA512):
		prf = sha512.New
	case kdfParams.Prf.Algorithm.Equal(oidHmacWithSHA1):
		prf = sha1.New
	case kdfParams.Prf.Algorithm.Equal(oidHmacWithMD5):
		prf = md5.New
	case kdfParams.Prf.Algorithm.Equal(asn1.ObjectIdentifier([]int{})):
		prf = sha1.New
	default:
		return nil, nil, NotImplementedError("pbes2 prf " + kdfParams.Prf.Algorithm.String() + " is not supported")
	}

	var keyLen int
	switch {
	case params.EncryptionScheme.Algorithm.Equal(oidAES256CBC):
		keyLen = 32
	case params.EncryptionScheme.Algorithm.Equal(oidAES192CBC):
		keyLen = 24
	case params.EncryptionScheme.Algorithm.Equal(oidAES128CBC):
		keyLen = 16
	default:
		return nil, nil, NotImplementedError("pbes2 algorithm " + params.EncryptionScheme.Algorithm.String() + " is not supported")
	}

	key := pbkdf2.Key(password, kdfParams.Salt.Bytes, kdfParams.Iterations, keyLen, prf)
	iv := params.EncryptionScheme.Parameters.Bytes

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	return block, iv, nil
}

// decryptable abstracts an object that contains ciphertext.
type decryptable interface {
	Algorithm() pkix.AlgorithmIdentifier
	Data() []byte
}

func pbEncrypterFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.BlockMode, int, error) {
	block, iv, err := pbeCipherFor(algorithm, password)
	if err != nil {
		return nil, 0, err
	}

	return cipher.NewCBCEncrypter(block, iv), block.BlockSize(), nil
}

func pbEncrypt(info encryptable, decrypted []byte, password []byte) error {
	cbc, blockSize, err := pbEncrypterFor(info.Algorithm(), password)
	if err != nil {
		return err
	}

	psLen := blockSize - len(decrypted)%blockSize
	encrypted := make([]byte, len(decrypted)+psLen)
	copy(encrypted[:len(decrypted)], decrypted)
	copy(encrypted[len(decrypted):], bytes.Repeat([]byte{byte(psLen)}, psLen))
	cbc.CryptBlocks(encrypted, encrypted)

	info.SetData(encrypted)

	return nil
}

// encryptable abstracts a object that contains ciphertext.
type encryptable interface {
	Algorithm() pkix.AlgorithmIdentifier
	SetData([]byte)
}

func makePBES2Parameters(rand io.Reader, salt []byte, iterations int) ([]byte, error) {
	var err error

	randomIV := make([]byte, 16)
	if _, err := rand.Read(randomIV); err != nil {
		return nil, err
	}

	var kdfparams pbkdf2Params
	if kdfparams.Salt.FullBytes, err = asn1.Marshal(salt); err != nil {
		return nil, err
	}
	kdfparams.Iterations = iterations
	kdfparams.Prf.Algorithm = oidHmacWithSHA256

	var params pbes2Params
	params.Kdf.Algorithm = oidPBKDF2
	if params.Kdf.Parameters.FullBytes, err = asn1.Marshal(kdfparams); err != nil {
		return nil, err
	}
	params.EncryptionScheme.Algorithm = oidAES256CBC
	if params.EncryptionScheme.Parameters.FullBytes, err = asn1.Marshal(randomIV); err != nil {
		return nil, err
	}

	return asn1.Marshal(params)
}

// PBKDF1 implementation for PKCS#5 PBE algorithms
func pbkdf1(hashFunc func() hash.Hash, password, salt []byte, iterationCount, dkLen int) []byte {
	if dkLen > 20 {
		// PBKDF1 can only generate up to hash length bytes
		return nil
	}

	h := hashFunc()
	h.Write(password)
	h.Write(salt)
	u := h.Sum(nil)

	for i := 1; i < iterationCount; i++ {
		h.Reset()
		h.Write(u)
		u = h.Sum(nil)
	}

	return u[:dkLen]
}

// Handle PKCS#5 PBE-SHA1-DES using PBKDF1
func pbkdf1CipherFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.Block, []byte, error) {
	var params pbeParams
	if err := unmarshal(algorithm.Parameters.FullBytes, &params); err != nil {
		return nil, nil, err
	}

	// Convert BMPString password to UTF-8 for PKCS#5
	originalPassword, err := decodeBMPString(password)
	if err != nil {
		return nil, nil, err
	}
	utf8Password := []byte(originalPassword)

	// PBKDF1 with SHA-1 to derive 16 bytes (8 for key + 8 for IV)
	derived := pbkdf1(sha1.New, utf8Password, params.Salt, params.Iterations, 16)
	if derived == nil {
		return nil, nil, errors.New("pbkdf1 derivation failed")
	}

	key := derived[:8]  // DES key is 8 bytes
	iv := derived[8:16] // IV is 8 bytes

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	return block, iv, nil
}
