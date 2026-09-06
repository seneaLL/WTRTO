package hud

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"strings"
	"time"

	"github.com/seneaLL/WTRTO/internal/i18n"
)

const ShareCodePrefixV1 = "WTRTO1-"

const ShareCodePrefix = "WTRTO2-"

type invalidShareCodeErr struct{}

func (invalidShareCodeErr) Error() string { return i18n.T("error.invalid_share_code") }

var ErrInvalidShareCode error = invalidShareCodeErr{}

func EncodeShareCode(t Template) (string, error) {
	data := encodeTemplateBinary(t)

	return wrapShareCode(ShareCodePrefix, data)
}

func DecodeShareCode(code string) (Template, error) {
	code = strings.TrimSpace(code)
	switch {
	case strings.HasPrefix(code, ShareCodePrefix):
		data, err := unwrapShareCode(ShareCodePrefix, code)
		if err != nil {
			return Template{}, err
		}
		t, err := decodeTemplateBinary(data)
		if err != nil {
			return Template{}, ErrInvalidShareCode
		}

		return t, nil
	case strings.HasPrefix(code, ShareCodePrefixV1):
		data, err := unwrapShareCode(ShareCodePrefixV1, code)
		if err != nil {
			return Template{}, err
		}
		var t Template
		if err := json.Unmarshal(data, &t); err != nil || t.Name == "" {
			return Template{}, ErrInvalidShareCode
		}

		return t, nil
	default:
		return Template{}, ErrInvalidShareCode
	}
}

func wrapShareCode(prefix string, data []byte) (string, error) {
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	if err := fw.Close(); err != nil {
		return "", err
	}
	compressed := buf.Bytes()

	out := make([]byte, 4+len(compressed))
	binary.BigEndian.PutUint32(out[:4], crc32.ChecksumIEEE(compressed))
	copy(out[4:], compressed)

	return prefix + base64.RawURLEncoding.EncodeToString(out), nil
}

func unwrapShareCode(prefix, code string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(code, prefix))
	if err != nil || len(raw) < 5 {
		return nil, ErrInvalidShareCode
	}
	want := binary.BigEndian.Uint32(raw[:4])
	compressed := raw[4:]
	if crc32.ChecksumIEEE(compressed) != want {
		return nil, ErrInvalidShareCode
	}
	fr := flate.NewReader(bytes.NewReader(compressed))
	defer fr.Close()
	data, err := io.ReadAll(fr)
	if err != nil {
		return nil, ErrInvalidShareCode
	}

	return data, nil
}

var kindByte = map[ElementKind]byte{KindText: 0, KindHorizon: 1, KindTapeV: 2, KindTapeH: 3}
var byteKind = map[byte]ElementKind{0: KindText, 1: KindHorizon, 2: KindTapeV, 3: KindTapeH}

const bindingNone = 0xFF

func bindingIndex(b Binding) byte {
	for i, ab := range AllBindings {
		if ab == b {
			return byte(i)
		}
	}

	return bindingNone
}

func bindingFromIndex(i byte) Binding {
	if int(i) < len(AllBindings) {
		return AllBindings[i]
	}

	return ""
}

func quantize01(v float64) uint16 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	return uint16(math.Round(v * 65535))
}

func dequantize01(q uint16) float64 {
	return float64(q) / 65535
}

type binWriter struct{ buf bytes.Buffer }

func (w *binWriter) byte(b byte) { w.buf.WriteByte(b) }
func (w *binWriter) u16(v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	w.buf.Write(b[:])
}
func (w *binWriter) uvarint(v uint64) {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(b[:], v)
	w.buf.Write(b[:n])
}
func (w *binWriter) svarint(v int64) {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutVarint(b[:], v)
	w.buf.Write(b[:n])
}
func (w *binWriter) str(s string)        { w.uvarint(uint64(len(s))); w.buf.WriteString(s) }
func (w *binWriter) frac(v float64)      { w.u16(quantize01(v)) }
func (w *binWriter) magnitude(v float64) { w.svarint(int64(math.Round(v * 10))) }
func (w *binWriter) color(c Color)       { w.buf.Write([]byte{c.R, c.G, c.B, c.A}) }

type binReader struct {
	r   *bytes.Reader
	err error
}

func (r *binReader) byte() byte {
	if r.err != nil {
		return 0
	}
	b, err := r.r.ReadByte()
	if err != nil {
		r.err = err
	}

	return b
}

func (r *binReader) u16() uint16 {
	if r.err != nil {
		return 0
	}
	var b [2]byte
	if _, err := io.ReadFull(r.r, b[:]); err != nil {
		r.err = err

		return 0
	}

	return binary.BigEndian.Uint16(b[:])
}

func (r *binReader) uvarint() uint64 {
	if r.err != nil {
		return 0
	}
	v, err := binary.ReadUvarint(r.r)
	if err != nil {
		r.err = err
	}

	return v
}

func (r *binReader) svarint() int64 {
	if r.err != nil {
		return 0
	}
	v, err := binary.ReadVarint(r.r)
	if err != nil {
		r.err = err
	}

	return v
}

func (r *binReader) str() string {
	n := r.uvarint()
	if r.err != nil || n > 1<<20 {
		return ""
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r.r, b); err != nil {
		r.err = err

		return ""
	}

	return string(b)
}

func (r *binReader) frac() float64 { return dequantize01(r.u16()) }

func (r *binReader) magnitude() float64 { return float64(r.svarint()) / 10 }

func (r *binReader) color() Color {
	if r.err != nil {
		return Color{}
	}
	var b [4]byte
	if _, err := io.ReadFull(r.r, b[:]); err != nil {
		r.err = err

		return Color{}
	}

	return Color{R: b[0], G: b[1], B: b[2], A: b[3]}
}

func encodeFlags(e Element) byte {
	var f byte
	if e.Bold {
		f |= 1 << 0
	}
	if e.Style == StyleArc {
		f |= 1 << 1
	}
	if e.Direction == DirDown || e.Direction == DirCCW {
		f |= 1 << 2
	}
	switch e.LabelSide {
	case SideLeft:
		f |= 1 << 3
	case SideRight:
		f |= 2 << 3
	}

	return f
}

func decodeFlags(f byte, kind ElementKind) (bold bool, style Style, direction Direction, side LabelSide) {
	bold = f&(1<<0) != 0
	if f&(1<<1) != 0 {
		style = StyleArc
	} else {
		style = StyleStraight
	}
	secondary := f&(1<<2) != 0
	if kind == KindTapeV {
		if secondary {
			direction = DirDown
		} else {
			direction = DirUp
		}
	} else {
		if secondary {
			direction = DirCCW
		} else {
			direction = DirCW
		}
	}
	switch (f >> 3) & 0x3 {
	case 1:
		side = SideLeft
	case 2:
		side = SideRight
	default:
		side = SideAuto
	}

	return
}

func encodeTemplateBinary(t Template) []byte {
	w := &binWriter{}
	w.str(t.Name)
	w.str(t.Army)
	w.uvarint(uint64(len(t.Elements)))
	for _, e := range t.Elements {
		w.byte(kindByte[e.Kind])
		w.byte(encodeFlags(e))
		switch e.Kind {
		case KindText:
			w.byte(bindingIndex(e.Binding))
			w.str(e.Label)
			w.str(e.Unit)
			w.frac(e.X)
			w.frac(e.Y)
			w.uvarint(uint64(e.FontSize))
			w.uvarint(uint64(e.Precision))
			w.color(e.Color)
		case KindHorizon:
			w.frac(e.X)
			w.frac(e.Y)
			w.frac(e.Size)
			w.color(e.Color)
		case KindTapeV, KindTapeH:
			w.byte(bindingIndex(e.Binding))
			w.frac(e.X)
			w.frac(e.Y)
			w.uvarint(uint64(e.FontSize))
			w.color(e.Color)
			w.frac(e.Length)
			w.magnitude(e.Range)
			w.magnitude(e.MinorStep)
			w.magnitude(e.MajorStep)
			w.magnitude(e.Wrap)
			w.uvarint(uint64(len(e.Zones)))
			for _, z := range e.Zones {
				w.magnitude(z.Threshold)
				w.color(z.Color)
			}
		}
	}

	return w.buf.Bytes()
}

func decodeTemplateBinary(data []byte) (Template, error) {
	r := &binReader{r: bytes.NewReader(data)}
	t := Template{}
	t.Name = r.str()
	t.Army = r.str()
	count := r.uvarint()
	if r.err != nil || t.Name == "" || count > 10000 {
		return Template{}, ErrInvalidShareCode
	}

	for i := uint64(0); i < count && r.err == nil; i++ {
		kb := r.byte()
		kind, ok := byteKind[kb]
		if !ok {
			return Template{}, ErrInvalidShareCode
		}
		flags := r.byte()
		bold, style, direction, side := decodeFlags(flags, kind)

		e := Element{
			ID:   fmt.Sprintf("el_%d_%d", time.Now().UnixNano(), i),
			Kind: kind,
			Bold: bold,
		}
		switch kind {
		case KindText:
			e.Binding = bindingFromIndex(r.byte())
			e.Label = r.str()
			e.Unit = r.str()
			e.X = r.frac()
			e.Y = r.frac()
			e.FontSize = int(r.uvarint())
			e.Precision = int(r.uvarint())
			e.Color = r.color()
		case KindHorizon:
			e.X = r.frac()
			e.Y = r.frac()
			e.Size = r.frac()
			e.Color = r.color()
		case KindTapeV, KindTapeH:
			e.Binding = bindingFromIndex(r.byte())
			e.X = r.frac()
			e.Y = r.frac()
			e.FontSize = int(r.uvarint())
			e.Color = r.color()
			e.Length = r.frac()
			e.Range = r.magnitude()
			e.MinorStep = r.magnitude()
			e.MajorStep = r.magnitude()
			e.Wrap = r.magnitude()
			e.Style = style
			e.Direction = direction
			e.LabelSide = side
			zc := r.uvarint()
			if zc > 1000 {
				return Template{}, ErrInvalidShareCode
			}
			for j := uint64(0); j < zc && r.err == nil; j++ {
				e.Zones = append(e.Zones, Zone{Threshold: r.magnitude(), Color: r.color()})
			}
		}
		t.Elements = append(t.Elements, e)
	}
	if r.err != nil {
		return Template{}, ErrInvalidShareCode
	}

	return t, nil
}
