package pavarotti

import (
	"github.com/viking/id3-go"
	"github.com/viking/id3-go/v2"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	tagVersion1 = iota
	tagVersion22
	tagVersion23
	tagVersionUnknown
)

var tagNameMap = map[string][]string{
	"Track":       {"", "TRK", "TRCK", ""},
	"AlbumArtist": {"", "TP2", "TPE2", ""},
	"Copyright":   {"", "TCR", "TCOP", ""},
	"Composer":    {"", "TCM", "TCOM", ""},
}

type Song struct {
	Path        string
	Artist      string
	AlbumArtist string
	Album       string
	Title       string
	Year        int
	Track       uint
	Genre       string
	Comments    []string
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

// Deduce song information from filename
func (song *Song) ParseBasename(basename string) {
	re := regexp.MustCompile("^(\\d+)\\s*-\\s*(.+)\\.[a-zA-Z0-9_]{3}$")
	matches := re.FindStringSubmatch(basename)
	if len(matches) == 3 {
		song.Title = matches[2]
		track, _ := strconv.ParseUint(matches[1], 10, 0)
		song.Track = uint(track)
	}
}

// Parse ID3 tags and update information
func (song *Song) UpdateFromMetadata() {
	var tagName string

	f, err := id3.Open(song.Path)
	if err != nil {
		return
	}
	defer f.Close()

	versionString := f.Version()
	var version uint8
	if versionString[:1] == "1" {
		version = tagVersion1
	} else if versionString[:3] == "2.2" {
		version = tagVersion22
	} else if versionString[:3] == "2.3" {
		version = tagVersion23
	} else {
		version = tagVersionUnknown
	}

	/// metadata that conditionally overwrites existing data
	artist := f.Artist()
	if artist != "" {
		song.Artist = artist
	}

	album := f.Album()
	if album != "" {
		song.Album = album
	}

	title := f.Title()
	if title != "" {
		song.Title = title
	}

	// track number
	var trackFrame v2.Framer
	tagName = tagNameMap["Track"][version]
	if tagName != "" {
		trackFrame = f.Frame(tagName)
	}
	if trackFrame != nil {
		track := trackFrame.String()
		if track != "" {
			var trackNum uint64
			trackNum, err = strconv.ParseUint(track, 10, 0)
			if err != nil {
				song.Track = uint(trackNum)
			}
		}
	}

	/// metadata that always overwrites existing data
	song.Genre = f.Genre()
	song.Year, _ = strconv.Atoi(f.Year())

	// comments
	comments := f.Comments()
	for _, comment := range comments {
		if comment != "" {
			song.Comments = append(song.Comments, comment)
		}
	}

	// album artist
	var albumArtistFrame v2.Framer
	tagName = tagNameMap["AlbumArtist"][version]
	if tagName != "" {
		albumArtistFrame = f.Frame(tagName)
	}
	if albumArtistFrame != nil {
		albumArtist := albumArtistFrame.String()
		if albumArtist != "" {
			song.AlbumArtist = albumArtist
		}
	}

	// copyright
	var copyrightFrame v2.Framer
	tagName = tagNameMap["Copyright"][version]
	if tagName != "" {
		copyrightFrame = f.Frame(tagName)
	}
	if copyrightFrame != nil {
		copyright := copyrightFrame.String()
		if copyright != "" {
			song.Copyright = copyright
		}
	}

	// composer
	var composerFrame v2.Framer
	tagName = tagNameMap["Composer"][version]
	if tagName != "" {
		composerFrame = f.Frame(tagName)
	}
	if composerFrame != nil {
		composer := composerFrame.String()
		if composer != "" {
			song.Composer = composer
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
