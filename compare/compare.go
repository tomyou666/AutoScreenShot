package compare

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/jpeg"
	"io"
)

// Hash は画像の JPEG エンコード後の SHA256 ハッシュを返します。
func Hash(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if quality <= 0 {
		quality = 85
	}
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	h := sha256.Sum256(buf.Bytes())
	return h[:], nil
}

// HashFromReader は読み込み済み画像データの SHA256 を返します（保存前のバイト列と一致させる用）。
func HashFromReader(r io.Reader) ([]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// ThreeSame は a, b, c の3つのハッシュがすべて一致するか返します。
func ThreeSame(a, b, c []byte) bool {
	if a == nil || b == nil || c == nil {
		return false
	}
	return bytes.Equal(a, b) && bytes.Equal(b, c)
}
