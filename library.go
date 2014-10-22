package pavarotti

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const DatabaseVersion = 1

type Library struct {
	DatabasePath string
	queue        []Song
	db           *sql.DB
}

func NewLibrary(databasePath string) (library *Library, err error) {
	library = &Library{DatabasePath: databasePath}
	err = library.openDatabase()
	return
}

func (library *Library) AddSong(song Song) {
	library.queue = append(library.queue, song)
	return
}

func (library *Library) Flush() (err error) {
	var (
		tx     *sql.Tx
		stmt1  *sql.Stmt
		stmt2  *sql.Stmt
		res    sql.Result
		lastId int64
	)

	tx, err = library.db.Begin()
	if err != nil {
		return
	}

	stmt1, err = tx.Prepare("INSERT INTO songs (path, title, artist, albumartist, album, track, year, genre, composer, copyright) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return
	}

	stmt2, err = tx.Prepare("INSERT INTO comments (data, song_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return
	}

	for _, song := range library.queue {
		res, err = stmt1.Exec(song.Path, song.Title, song.Artist, song.AlbumArtist, song.Album, song.Track, song.Year, song.Genre, song.Composer, song.Copyright)
		if err != nil {
			tx.Rollback()
			return
		}

		if len(song.Comments) > 0 {
			lastId, err = res.LastInsertId()
			if err != nil {
				tx.Rollback()
				return
			}

			for _, comment := range song.Comments {
				_, err = stmt2.Exec(comment, lastId)
				if err != nil {
					tx.Rollback()
					return
				}
			}
		}
	}

	err = tx.Commit()
	library.queue = nil
	return
}

func (library *Library) Close() (err error) {
	if library.db != nil {
		err = library.Flush()
		library.db.Close()
		library.db = nil
	}
	return
}

func (library *Library) openDatabase() (err error) {
	var (
		rows    *sql.Rows
		version uint
	)

	library.db, err = sql.Open("sqlite3", library.DatabasePath)
	if err != nil {
		return
	}

	rows, err = library.db.Query("SELECT version FROM schema_info")
	if err == nil && rows.Next() {
		err = rows.Scan(&version)
		if err != nil {
			library.db.Close()
			return
		}
	}

	for version < DatabaseVersion {
		switch version {
		case 0:
			_, err = library.db.Exec("CREATE TABLE songs (id INTEGER PRIMARY KEY, path TEXT, title TEXT, artist TEXT, albumartist TEXT, album TEXT, track INTEGER, year INTEGER, genre TEXT, composer TEXT, copyright TEXT)")
			if err != nil {
				library.db.Close()
				return
			}

			_, err = library.db.Exec("CREATE TABLE comments (id INTEGER PRIMARY KEY, data TEXT, song_id INTEGER)")
			if err != nil {
				library.db.Close()
				return
			}

			_, err = library.db.Exec("CREATE TABLE schema_info (version INTEGER)")
			if err != nil {
				library.db.Close()
				return
			}

			_, err = library.db.Exec("INSERT INTO schema_info (version) VALUES (1)")
			if err != nil {
				library.db.Close()
				return
			}
		}
		version += 1
	}

	return
}
