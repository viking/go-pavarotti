package pavarotti

import (
	"os"
	"path/filepath"
)

func NewSongFromPath(root, relPath string) Song {
	songFile := SongFile{root, relPath}
	return songFile.CreateSong()
}

func Find(root string, ch chan Song) error {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			song := NewSongFromPath(root, relPath)
			song.Path = path
			ch <- song
		}
		return nil
	}
	return filepath.Walk(root, walkFunc)
}
