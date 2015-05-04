package drum

import (
  "os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
  file, err := os.Open(path)
  defer file.Close()
  if err != nil {
    return nil, err
  }

  return Decode(file)
}
