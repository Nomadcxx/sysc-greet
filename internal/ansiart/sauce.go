package ansiart

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// Sauce is a SAUCE 00.5 metadata record as stored after an .ans body.
type Sauce struct {
	Title, Author, Group, Date string
	FileSize                   uint32
	DataType, FileType         uint8
	TInfo1, TInfo2             uint16
	TInfo3, TInfo4             uint16
	Flags                      uint8
	TInfoS                     string
	Comments                   []string
}

// ICE reports whether the iCE colours flag is set (Flags bit 0).
func (s *Sauce) ICE() bool { return s != nil && s.Flags&1 != 0 }

func cp437String(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		sb.WriteRune(CP437[c])
	}
	return strings.TrimRight(sb.String(), " \u0000")
}

// ParseSAUCE splits b into the file body and its SAUCE record. It is
// tolerant: without a SAUCE record it returns (b, nil), stripping a
// trailing 0x1A EOF marker when present.
func ParseSAUCE(b []byte) ([]byte, *Sauce) {
	if len(b) < 128 || !bytes.Equal(b[len(b)-128:len(b)-123], []byte("SAUCE")) {
		if i := bytes.IndexByte(b, 0x1a); i >= 0 {
			return b[:i], nil
		}
		return b, nil
	}
	r := b[len(b)-128:]
	s := &Sauce{
		Title:  cp437String(r[7:42]),
		Author: cp437String(r[42:62]),
		Group:  cp437String(r[62:82]),
		Date:   string(r[82:90]),
	}
	s.FileSize = binary.LittleEndian.Uint32(r[90:94])
	s.DataType, s.FileType = r[94], r[95]
	s.TInfo1 = binary.LittleEndian.Uint16(r[96:98])
	s.TInfo2 = binary.LittleEndian.Uint16(r[98:100])
	s.TInfo3 = binary.LittleEndian.Uint16(r[100:102])
	s.TInfo4 = binary.LittleEndian.Uint16(r[102:104])
	s.Flags = r[105]
	s.TInfoS = strings.TrimRight(string(r[106:128]), "\x00")
	body := b[:len(b)-128]
	if n := int(r[104]); n > 0 && len(body) >= 5+64*n {
		c := body[len(body)-5-64*n:]
		if bytes.HasPrefix(c, []byte("COMNT")) {
			for i := 0; i < n; i++ {
				s.Comments = append(s.Comments, cp437String(c[5+64*i:5+64*(i+1)]))
			}
			body = body[:len(body)-5-64*n]
		}
	}
	if i := bytes.IndexByte(body, 0x1a); i >= 0 {
		body = body[:i]
	}
	return body, s
}

// EncodeSAUCE builds a SAUCE00 record with an optional COMNT block.
func EncodeSAUCE(s Sauce) []byte {
	pad := func(v string, n int) []byte {
		out := bytes.Repeat([]byte{' '}, n)
		copy(out, v)
		return out
	}
	var buf bytes.Buffer
	buf.WriteByte(0x1a)
	if len(s.Comments) > 0 {
		buf.WriteString("COMNT")
		for _, c := range s.Comments {
			buf.Write(pad(c, 64))
		}
	}
	buf.WriteString("SAUCE00")
	buf.Write(pad(s.Title, 35))
	buf.Write(pad(s.Author, 20))
	buf.Write(pad(s.Group, 20))
	buf.Write(pad(s.Date, 8))
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, s.FileSize)
	buf.Write(tmp)
	buf.WriteByte(s.DataType)
	buf.WriteByte(s.FileType)
	t := make([]byte, 2)
	binary.LittleEndian.PutUint16(t, s.TInfo1)
	buf.Write(t)
	binary.LittleEndian.PutUint16(t, s.TInfo2)
	buf.Write(t)
	binary.LittleEndian.PutUint16(t, s.TInfo3)
	buf.Write(t)
	binary.LittleEndian.PutUint16(t, s.TInfo4)
	buf.Write(t)
	buf.WriteByte(byte(len(s.Comments)))
	buf.WriteByte(s.Flags)
	f := make([]byte, 22)
	copy(f, s.TInfoS)
	buf.Write(f)
	return buf.Bytes()
}
