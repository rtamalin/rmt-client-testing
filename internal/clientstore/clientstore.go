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

type FileType string

const (
	SYS_INFO_TYPE FileType = "sysinfo"
	REG_INFO_TYPE FileType = "reginfo"
)

type FileId uint32

func (i FileId) topDirBits() uint32 {
	return uint32(i) >> 20
}

func (i FileId) midDirBits() uint32 {
	return (uint32(i) >> 10) & 0x3ff
}

func (i FileId) leafDirBits() uint32 {
	return uint32(i) & 0x3ff
}

func (i FileId) DirPath() string {
	return filepath.Join(
		fmt.Sprintf("%03x", i.topDirBits()),
		fmt.Sprintf("%03x", i.midDirBits()),
		fmt.Sprintf("%03x", i.leafDirBits()),
	)
}

func (i FileId) FileName(fileType FileType) string {
	return fmt.Sprintf("%s.json", fileType)
}

func (i FileId) Path(fileType FileType) string {
	return filepath.Join(
		i.DirPath(),
		i.FileName(fileType),
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

func (s *ClientStore) ClientDirPath(id FileId) string {
	return filepath.Join(s.rootDir, id.DirPath())
}

func (s *ClientStore) ClientPath(id FileId, fileType FileType) string {
	return filepath.Join(s.rootDir, id.Path(fileType))
}

func (s *ClientStore) EnsureDirectoryExists(id FileId) (err error) {
	dirPath := s.ClientDirPath(id)

	// ensure directory hierarchy exists for file
	if err = os.MkdirAll(dirPath, 0o755); err != nil {
		err = fmt.Errorf(
			"failed to create datastore file hierarchy %q: %w",
			dirPath,
			err,
		)
		return
	}

	return
}

func (s *ClientStore) WriteFile(id FileId, fileType FileType, data []byte, perm os.FileMode) (err error) {
	if err = s.EnsureDirectoryExists(id); err != nil {
		err = fmt.Errorf(
			"failed to write datastore file: %w",
			err,
		)
	}

	filePath := s.ClientPath(id, fileType)
	if err = os.WriteFile(filePath, data, perm); err != nil {
		err = fmt.Errorf(
			"failed to write datastore file %q: %w",
			filePath,
			err,
		)
		// delete any partially created file, ignoring the error
		_ = s.Delete(id, fileType)
		return
	}

	return
}

func (s *ClientStore) ReadFile(id FileId, fileType FileType) (data []byte, err error) {
	filePath := s.ClientPath(id, fileType)

	if data, err = os.ReadFile(filePath); err != nil {
		err = fmt.Errorf(
			"failed to read datastore file %q: %w",
			filePath,
			err,
		)
		return
	}

	return
}

func (s *ClientStore) Delete(id FileId, fileType FileType) (err error) {
	filePath := s.ClientPath(id, fileType)

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

func (s *ClientStore) Exists(id FileId, fileType FileType) bool {
	filePath := s.ClientPath(id, fileType)

	st, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// doesn't exist
			return false
		}

		// some other error, possibly path access permissions
		return false
	}
	if st.IsDir() {
		// path does exist but is a directory, but doesn't matter for now
		return true
	}

	return true
}
