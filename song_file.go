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

type SongFile struct {
	RootPath string
	RelPath  string
}

var basenamePattern = regexp.MustCompile("^(\\d+)\\s*-\\s*(.+)\\.[a-zA-Z0-9_]{3}$")

type songPathInfo struct {
	artist string
	album  string
	title  string
	track  uint64
}

type songID3Info struct {
	artist      string
	album       string
	title       string
	track       uint64
	genre       string
	year        int
	comments    []string
	albumArtist string
	copyright   string
	composer    string
}

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

func (sf SongFile) CreateSong() (song Song) {
	pathInfo := sf.pathInfo()
	song.Artist = pathInfo.artist
	song.Album = pathInfo.album
	song.Title = pathInfo.title
	song.Track = uint(pathInfo.track)

	id3Info, err := sf.id3Info()
	if err != nil {
		return
	}
	if id3Info.artist != "" {
		song.Artist = id3Info.artist
	}
	if id3Info.album != "" {
		song.Album = id3Info.album
	}
	if id3Info.title != "" {
		song.Title = id3Info.title
	}
	if id3Info.track != 0 {
		song.Track = uint(id3Info.track)
	}
	song.Genre = id3Info.genre
	song.Year = id3Info.year
	song.Comments = id3Info.comments
	song.AlbumArtist = id3Info.albumArtist
	song.Copyright = id3Info.copyright
	song.Composer = id3Info.composer

	return
}

func (sf SongFile) pathInfo() (info songPathInfo) {
	parts := strings.Split(sf.RelPath, string(os.PathSeparator))
	plen := len(parts)
	if plen > 2 {
		info.artist = parts[plen-3]
		info.album = parts[plen-2]
	} else if plen > 1 {
		info.artist = parts[plen-2]
	}

	matches := basenamePattern.FindStringSubmatch(parts[plen-1])
	if len(matches) == 3 {
		info.title = matches[2]
		info.track, _ = strconv.ParseUint(matches[1], 10, 0)
	}
	return
}

func (sf SongFile) id3Info() (info songID3Info, err error) {
	var tagName string
	var frame v2.Framer
	var f *id3.File

	f, err = id3.Open(filepath.Join(sf.RootPath, sf.RelPath))
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

	info.artist = f.Artist()
	info.album = f.Album()
	info.title = f.Title()
	info.genre = f.Genre()
	info.year, _ = strconv.Atoi(f.Year())

	// track
	tagName = tagNameMap["Track"][version]
	if tagName != "" {
		if frame = f.Frame(tagName); frame != nil {
			if track := frame.String(); track != "" {
				info.track, _ = strconv.ParseUint(track, 10, 0)
			}
		}
	}

	// comments
	comments := f.Comments()
	for _, comment := range comments {
		if comment != "" {
			info.comments = append(info.comments, comment)
		}
	}

	// album artist
	tagName = tagNameMap["AlbumArtist"][version]
	if tagName != "" {
		if frame = f.Frame(tagName); frame != nil {
			info.albumArtist = frame.String()
		}
	}

	// copyright
	tagName = tagNameMap["Copyright"][version]
	if tagName != "" {
		if frame = f.Frame(tagName); frame != nil {
			info.copyright = frame.String()
		}
	}

	// composer
	tagName = tagNameMap["Composer"][version]
	if tagName != "" {
		if frame = f.Frame(tagName); frame != nil {
			info.composer = frame.String()
		}
	}

	return
}
