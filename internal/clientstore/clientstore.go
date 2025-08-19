package clientstore

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

type FileId uint32

func (i FileId) topDirBits() uint32 {
	return uint32(i) >> 20
}

func (i FileId) midDirBits() uint32 {
	return (uint32(i) >> 10) & 0x3ff
}

func (i FileId) leafBits() uint32 {
	return uint32(i) & 0x3ff
}

func (i FileId) DirPath() string {
	return filepath.Join(
		fmt.Sprintf("%03x", i.topDirBits()),
		fmt.Sprintf("%03x", i.midDirBits()),
	)
}

func (i FileId) FileName() string {
	return fmt.Sprintf("%03x.json", i.leafBits())
}

func (i FileId) Path() string {
	return filepath.Join(
		i.DirPath(),
		i.FileName(),
	)
}

type ClientStore struct {
	rootDir string
}

func New(rootPath string) *ClientStore {
	s := new(ClientStore)
	s.Init(rootPath)
	return s
}

func (s *ClientStore) Init(rootPath string) {
	fi, err := os.Stat(rootPath)

	// fail if we don't have permissions to access the path
	if err != nil && errors.Is(err, fs.ErrPermission) {
		log.Fatalf(
			"Permission denied for datastore root %q: %s",
			rootPath,
			err.Error(),
		)
	}

	// if rootPath doesn't exist, create it
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		// create the root directory
		err = os.MkdirAll(rootPath, 0o755)
		if err != nil {
			log.Fatalf(
				"Failed to os.MkdirAll() datastore root %q: %s",
				rootPath,
				err.Error(),
			)
		}

		// stat the newly created root
		fi, err = os.Stat(rootPath)
	}

	// os.Stat() failed for some other reason
	if err != nil {
		log.Fatalf(
			"Failed to os.Stat() datastore root %q: %s",
			rootPath,
			err.Error(),
		)
	}

	// specified rootPath must be a directory with viable access
	if !fi.IsDir() {
		log.Fatalf(
			"Specified datastore root %q is not a directory",
			rootPath,
		)
	}
	if err = unix.Access(rootPath, unix.W_OK|unix.X_OK); err != nil {
		log.Fatalf(
			"Specified datastore root %q lacks appropriate access permissions: %s",
			rootPath,
			err.Error(),
		)
	}

	s.rootDir = rootPath
}

func (s *ClientStore) String() string {
	return fmt.Sprintf(
		"{root:%q}",
		s.rootDir,
	)
}

func (s *ClientStore) Root() string {
	return s.rootDir
}

func (s *ClientStore) Open(id FileId, create bool) (fp *os.File, err error) {
	dirPath := filepath.Join(s.rootDir, id.DirPath())
	filePath := filepath.Join(s.rootDir, id.Path())

	// ensure directory hierarchy exists for file
	if err = os.MkdirAll(dirPath, 0o755); err != nil {
		err = fmt.Errorf(
			"failed to create datastore file hierarchy %q: %w",
			dirPath,
			err,
		)
		return
	}

	// open file, creating if necessary
	flags := os.O_RDWR
	action := "open"
	if create {
		flags |= os.O_CREATE
		action = "create"
	}
	fp, err = os.OpenFile(filePath, flags, 0o644)
	if err != nil {
		err = fmt.Errorf(
			"failed to %s datastore file %q: %w",
			action,
			filePath,
			err,
		)
		return
	}

	return
}

func (s *ClientStore) Delete(id FileId) (err error) {
	filePath := filepath.Join(s.rootDir, id.Path())

	err = os.Remove(filePath)
	if err != nil {
		err = fmt.Errorf(
			"failed to delete datastore file %q: %w",
			filePath,
			err,
		)
		return
	}

	return
}
