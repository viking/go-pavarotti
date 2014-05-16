package pavarotti

import (
	"github.com/mikkyang/id3-go"
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

func NewSongFromPath(root, relPath string) (song Song) {
	var songPath, songBase string

	songPath = filepath.Join(root, relPath)
	song.Path = songPath

	// deduce artist and album from path if possible
	parts := strings.Split(relPath, string(os.PathSeparator))
	lp := len(parts)
	songBase = parts[lp-1]
	if lp > 2 {
		song.Album = parts[lp-2]
		song.Artist = parts[lp-3]
	} else if lp > 1 {
		song.Artist = parts[lp-2]
	}

	song.ParseBasename(songBase)
	song.UpdateFromMetadata()
	return
}

func (Song) sanitize(s string) string {
	// strip null bytes
	i := strings.Index(s, "\x00")
	if i >= 0 {
		return s[:i]
	}
	return s
}

// Deduce song information from filename
func (song *Song) ParseBasename(basename string) {
	re := regexp.MustCompile("^(\\d+)\\s*-\\s*(.+)\\.[a-zA-Z0-9_]{3}$")
	matches := re.FindStringSubmatch(basename)
	if len(matches) == 3 {
		song.Title = matches[2]
		song.Track = matches[1]
	}
}

// Parse ID3 tags and update information
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
