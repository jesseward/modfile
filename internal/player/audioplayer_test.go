package player

import (
	"bytes"
	"io"
	"testing"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func TestWavPlayer_writeWavHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	player := NewWavPlayer(nopCloser{buf}, DefaultPlayerOptions())
	player.writeWavHeader()

	// RIFF header
	expected := []byte("RIFF")
	if !bytes.Equal(buf.Bytes()[:4], expected) {
		t.Errorf("expected RIFF header, got %v", buf.Bytes()[:4])
	}

	// WAVE format
	expected = []byte("WAVE")
	if !bytes.Equal(buf.Bytes()[8:12], expected) {
		t.Errorf("expected WAVE format, got %v", buf.Bytes()[8:12])
	}

	// "fmt " sub-chunk
	expected = []byte("fmt ")
	if !bytes.Equal(buf.Bytes()[12:16], expected) {
		t.Errorf("expected 'fmt ' sub-chunk, got %v", buf.Bytes()[12:16])
	}

	// "data" sub-chunk
	expected = []byte("data")
	if !bytes.Equal(buf.Bytes()[36:40], expected) {
		t.Errorf("expected 'data' sub-chunk, got %v", buf.Bytes()[36:40])
	}
}
