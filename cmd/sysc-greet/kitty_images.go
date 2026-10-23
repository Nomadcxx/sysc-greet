package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"strings"
	"sync"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
)

const (
	kittyGraphicsChunkSize = 4096
	kittyBackgroundZ       = -2147483648
)

type kittyImagePresenter struct {
	ids       [2]uint32
	next      int
	currentID uint32
	out       *kittyOutput
}

// kittyOutput keeps graphics writes in update order. Bubble Tea runs each
// command on its own goroutine, so tea.Raw writes from separate updates can
// reach the terminal out of order and strand an image after its delete. The
// event loop prints a RawMsg with fmt.Sprint, so the queue drains there: the
// first write to arrive carries every sequence queued so far.
type kittyOutput struct {
	mu      sync.Mutex
	pending strings.Builder
}

func (o *kittyOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	sequence := o.pending.String()
	o.pending.Reset()
	return sequence
}

func (p *kittyImagePresenter) write(sequence string) tea.Cmd {
	if sequence == "" {
		return nil
	}
	if p.out == nil {
		p.out = &kittyOutput{}
	}
	p.out.mu.Lock()
	p.out.pending.WriteString(sequence)
	p.out.mu.Unlock()
	return tea.Raw(p.out)
}

func kittyGraphicsAvailable() bool {
	return os.Getenv("KITTY_WINDOW_ID") != ""
}

func (p *kittyImagePresenter) render(frame syscgoanimations.MeshFrame, columns, rows int) (string, error) {
	if frame.Pixels == nil || columns <= 0 || rows <= 0 {
		return "", fmt.Errorf("invalid Kitty mesh frame dimensions")
	}
	if p.ids[0] == 0 || p.ids[1] == 0 {
		base := uint32(os.Getpid()) << 1
		if base == 0 {
			base = 2
		}
		p.ids = [2]uint32{base, base + 1}
	}

	var pngData bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(&pngData, frame.Pixels); err != nil {
		return "", fmt.Errorf("encode Kitty mesh frame: %w", err)
	}
	imageID := p.ids[p.next]
	sequence, err := kittyPNGCommands(pngData.Bytes(), imageID, columns, rows, p.currentID)
	if err != nil {
		return "", err
	}
	p.currentID = imageID
	p.next = 1 - p.next
	return sequence, nil
}

func (p *kittyImagePresenter) clear() string {
	if p.currentID == 0 {
		return ""
	}
	// Delete both IDs: a frame queued but never written leaves the other on screen.
	var sequence strings.Builder
	for _, id := range p.ids {
		if id != 0 {
			fmt.Fprintf(&sequence, "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", id)
		}
	}
	p.currentID = 0
	return sequence.String()
}

func kittyPNGCommands(pngData []byte, imageID uint32, columns, rows int, deleteID uint32) (string, error) {
	if len(pngData) == 0 || imageID == 0 || columns <= 0 || rows <= 0 {
		return "", fmt.Errorf("invalid Kitty image command")
	}
	encoded := base64.StdEncoding.EncodeToString(pngData)
	var sequence strings.Builder
	sequence.Grow(len(encoded) + len(encoded)/kittyGraphicsChunkSize*40 + 160)
	sequence.WriteString("\x1b[H")
	for start := 0; start < len(encoded); start += kittyGraphicsChunkSize {
		end := min(start+kittyGraphicsChunkSize, len(encoded))
		more := 0
		if end < len(encoded) {
			more = 1
		}
		if start == 0 {
			fmt.Fprintf(&sequence, "\x1b_Ga=T,f=100,i=%d,p=1,c=%d,r=%d,z=%d,C=1,q=2,m=%d;%s\x1b\\", imageID, columns, rows, kittyBackgroundZ, more, encoded[start:end])
		} else {
			fmt.Fprintf(&sequence, "\x1b_Gm=%d,q=2;%s\x1b\\", more, encoded[start:end])
		}
	}
	if deleteID != 0 && deleteID != imageID {
		fmt.Fprintf(&sequence, "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", deleteID)
	}
	return sequence.String(), nil
}
