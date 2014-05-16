package pavarotti

import (
	"github.com/mikkyang/id3-go"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Song struct {
	Path        string
	Artist      string
	AlbumArtist string
	Album       string
	Title       string
	Year        string
	Track       string
	Genre       string
	Comment     string
	Composer    string
	Copyright   string
}

func (Song) sanitize(s string) string {
	// strip null bytes
	i := strings.Index(s, "\x00")
	if i >= 0 {
		return s[:i]
	}
	return s
}

func (song *Song) UpdateFromMetadata() {
	f, err := id3.Open(song.Path)
	if err != nil {
		return
	}
	defer f.Close()

	artist := song.sanitize(f.Artist())
	if artist != "" {
		song.Artist = artist
	}

	album := song.sanitize(f.Album())
	if album != "" {
		song.Album = album
	}

	title := song.sanitize(f.Title())
	if title != "" {
		song.Title = title
	}

	trackFrame := f.Frame("TRCK")
	if trackFrame != nil {
		track := song.sanitize(trackFrame.String())
		if track != "" {
			song.Track = track
		}
	}
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
				song := Song{
					Path:   filepath.Join(path, file.Name()),
					Artist: artistName,
					Album:  albumName,
					Title:  matches[2],
					Track:  matches[1],
				}
				song.UpdateFromMetadata()
				ch <- song
			}
		}
	}
	return
}
