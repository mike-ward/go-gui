// Package tempfont creates unique temporary font files for
// backends that must hand a persistent path to font loaders.
package tempfont

import "os"

// Write creates a unique temporary font file containing data
// and returns its path.
func Write(prefix string, data []byte) (string, error) {
	f, err := os.CreateTemp("", prefix+"-*.ttf")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}
