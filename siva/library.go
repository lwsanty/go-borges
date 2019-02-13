package siva

import (
	"fmt"
	"io"
	"os"
	"strings"

	borges "github.com/src-d/go-borges"
	"github.com/src-d/go-borges/util"
	billy "gopkg.in/src-d/go-billy.v4"
	butil "gopkg.in/src-d/go-billy.v4/util"
)

// Library represents a borges.Library implementation based on siva files.
type Library struct {
	id borges.LibraryID
	fs billy.Filesystem
}

var _ borges.Library = (*Library)(nil)

// NewLibrary creates a new siva.Library.
func NewLibrary(id string, fs billy.Filesystem) *Library {
	return &Library{
		id: borges.LibraryID(id),
		fs: fs,
	}
}

// ID implements borges.Library interface.
func (l *Library) ID() borges.LibraryID {
	return l.id
}

// Init implements borges.Library interface.
func (l *Library) Init(borges.RepositoryID) (borges.Repository, error) {
	return nil, borges.ErrNotImplemented.New()
}

// Get implements borges.Library interface.
func (l *Library) Get(repoID borges.RepositoryID, mode borges.Mode) (borges.Repository, error) {
	ok, _, locID, err := l.Has(repoID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, borges.ErrRepositoryNotExists.New(repoID)
	}

	loc, err := l.Location(locID)
	if err != nil {
		return nil, err
	}

	return loc.Get(repoID, mode)
}

// GetOrInit implements borges.Library interface.
func (l *Library) GetOrInit(borges.RepositoryID) (borges.Repository, error) {
	return nil, borges.ErrNotImplemented.New()
}

// TODO: find if we have to use ".git" suffix for repository ids
func toRepoID(endpoint string) borges.RepositoryID {
	name, _ := borges.NewRepositoryID(endpoint)
	return borges.RepositoryID(strings.TrimSuffix(name.String(), ".git"))
}

func toLocID(file string) borges.LocationID {
	id := strings.TrimSuffix(file, ".siva")
	return borges.LocationID(id)
}

// Has implements borges.Library interface.
func (l *Library) Has(name borges.RepositoryID) (bool, borges.LibraryID, borges.LocationID, error) {
	it, err := l.Locations()
	if err != nil {
		return false, "", "", err
	}
	defer it.Close()

	for {
		loc, err := it.Next()
		if err == io.EOF {
			return false, "", "", nil
		}
		if err != nil {
			return false, "", "", err
		}

		has, err := loc.Has(name)
		if err != nil {
			return false, "", "", err
		}

		if has {
			return true, l.id, loc.ID(), nil
		}
	}
}

// Repositories implements borges.Library interface.
func (l *Library) Repositories(mode borges.Mode) (borges.RepositoryIterator, error) {
	locs, err := l.locations()
	if err != nil {
		return nil, err
	}

	return util.NewLocationRepositoryIterator(locs, mode), nil
}

// Location implements borges.Library interface.
func (l *Library) Location(id borges.LocationID) (borges.Location, error) {
	return l.generateLocation(id)
}

func (l *Library) generateLocation(id borges.LocationID) (*Location, error) {
	path := fmt.Sprintf("%s.siva", id)
	_, err := l.fs.Stat(path)
	if os.IsNotExist(err) {
		return nil, borges.ErrLocationNotExists.New(id)
	}

	return NewLocation(id, l, path)
}

// Locations implements borges.Library interface.
func (l *Library) Locations() (borges.LocationIterator, error) {
	locs, err := l.locations()
	if err != nil {
		return nil, err
	}

	return util.NewLocationIterator(locs), nil
}

func (l *Library) locations() ([]borges.Location, error) {
	var locs []borges.Location

	sivas, err := butil.Glob(l.fs, "*.siva")
	if err != nil {
		return nil, err
	}

	for _, s := range sivas {
		loc, err := l.generateLocation(toLocID(s))
		if err != nil {
			continue
		}

		locs = append(locs, loc)
	}

	return locs, nil
}

// Library implements borges.Library interface.
func (l *Library) Library(id borges.LibraryID) (borges.Library, error) {
	if id == l.id {
		return l, nil
	}

	return nil, borges.ErrLibraryNotExists.New(id)
}

// Libraries implements borges.Library interface.
func (l *Library) Libraries() (borges.LibraryIterator, error) {
	libs := []borges.Library{l}
	return util.NewLibraryIterator(libs), nil
}
