// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
  "os"
  "encoding/binary"
  "bytes"
  "fmt"
  "math"
  "errors"
)

/*
FILE STRUCTURE:
SPLICE HEADER   - 6   bytes
buffer          - 7   bytes
VERSION         - 11  bytes ->
buffer          - 21  bytes -> 36 bytes
TEMPO           - 4   bytes ->
TRACK:
  ID            - 1   byte
  buffer        - 3   bytes
  NAME LEN      - 1   byte
  NAME          - x   bytes
  BEAT          - 16  bytes
*/

var spliceHeaderLength  int64 = 14
var versionLength       int64 = 32
var tempoLength         int64 = 4
var trackIdLength       int64 = 1
var errBytesLength      int64 = 3
var trackLength         int64 = 16

func Decode(f *os.File) (*Pattern, error) {
  var offset int64

  offset, spliceHeader := readAndSeek(f, spliceHeaderLength, 0)

  if bytes.Index(spliceHeader, []byte{'S', 'P', 'L', 'I', 'C', 'E'}) == -1 {
    return nil, errors.New("File does not contain correct header.")
  }

  // Grab the version and trim trailing 0x00s
  offset, version := readAndSeek(f, versionLength, offset)
  version = trim(version)

  // Grab the float32 tempo
  var tempo float32
  offset, tempoBuf := readAndSeek(f, tempoLength, offset)
  buffer := bytes.NewReader(tempoBuf)
  binary.Read(buffer, binary.LittleEndian, &tempo)

  fileInfo, _ := f.Stat()
  fileLength := fileInfo.Size()

  dPatterns := []DrumPattern{}
  // Loop through the file while we are still within the confines of the length
  for offset, _ = f.Seek(0, 1); offset < fileLength; offset, _ = f.Seek(0, 1) {
    // First byte is the track id
    offset, id := readAndSeek(f, trackIdLength, offset)

    // If not followed by three 0 bytes, this file is *corrupted*, stop here
    offset, errBytes := readAndSeek(f, errBytesLength, offset)
    if bytes.Compare(errBytes, []byte{0, 0, 0}) != 0 {
      break;
    }

    // The fifth byte lets us know how long the name of the track is
    offset, nameLength := readAndSeek(f, 1, offset)
    offset, name := readAndSeek(f, int64(nameLength[0]), offset)
    offset, track := readAndSeek(f, trackLength, offset)

    dPatterns = append(dPatterns, DrumPattern{id: id[0], name: name, track: track})
  }

  p := &Pattern{version: version, tempo: tempo, patterns: dPatterns}

  return p, nil
}

type DrumPattern struct {
  id     byte
  name   []byte
  track  []byte
}

type Pattern struct {
  version  []byte
  tempo    float32
  patterns []DrumPattern
}

func (p Pattern) Format(f fmt.State, c rune) {
  fmt.Fprintf(f, "Saved with HW Version: %s\n", p.version)
  fmt.Fprintf(f, "Tempo: %v\n", p.tempo)
  for _, pattern := range p.patterns {
    fmt.Fprintf(f, "(%v) %s\t%s\n", pattern.id, pattern.name, buildpattern(pattern.track))
  }
}

func buildpattern(pattern []byte) (n []byte) {
  var newPattern []byte
  // Start with a '|' character
  newPattern = append(newPattern, 0x7c)

  for key, val := range pattern {
    // If the value is one, append 'x', else append '-'
    if val == 1 {
      newPattern = append(newPattern, 0x78)
    } else {
      newPattern = append(newPattern, 0x2d)
    }

    // Add a '|' every 4 characters
    if math.Mod(float64(key), 4) == 3 {
      newPattern = append(newPattern, 0x7c)
    }
  }
  return newPattern
}

// Return a slice that does not contain trailing 0s
func trim(b []byte) []byte {
  if i := bytes.IndexByte(b, 0x00); i != -1 {
    b = b[:i]
  }
  return b
}

// Take in a file, length and offset to begin to read at
// Returns the new offset and the byte array
func readAndSeek(f *os.File, length, ofst int64) (off int64, buffer []byte) {
  b := make([]byte, length)
  f.ReadAt(b, ofst)
  o, _ := f.Seek(length, 1)
  return o, b
}
