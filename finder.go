package pavarotti

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type Song struct {
	Path        string
	Artist      string
	AlbumArtist string
	Album       string
	Title       string
	Year        uint
	Track       uint
	Genre       string
	Comment     string
	Composer    string
	Copyright   string
}

func Find(path string, ch chan Song) (err error) {
	var files []os.FileInfo
	files, err = ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			artistName := file.Name()
			subPath := filepath.Join(path, artistName)
			err = walkArtist(subPath, artistName, ch)
			if err != nil {
				return
			}
		}
	}
	return
}

func walkArtist(path string, artistName string, ch chan Song) (err error) {
	var files []os.FileInfo
	files, err = ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			albumName := file.Name()
			subPath := filepath.Join(path, albumName)
			err = walkAlbum(subPath, artistName, albumName, ch)
			if err != nil {
				return
			}
		}
	}
	return
}

func walkAlbum(path string, artistName string, albumName string, ch chan Song) (err error) {
	var files []os.FileInfo
	files, err = ioutil.ReadDir(path)
	if err != nil {
		return
	}

	re := regexp.MustCompile("^(\\d+)\\s*-\\s*(.+)\\.[a-zA-Z0-9_]{3}$")
	for _, file := range files {
		if !file.IsDir() {
			matches := re.FindStringSubmatch(file.Name())
			if len(matches) == 3 {
				var track uint64
				song := Song{
					Path:   filepath.Join(path, file.Name()),
					Artist: artistName,
					Album:  albumName,
					Title:  matches[2],
				}
				track, err = strconv.ParseUint(matches[1], 10, 0)
				if err == nil {
					song.Track = uint(track)
				}
				ch <- song
			}
		}
	}
	return
}
