package ssh

import (
	"errors"
	"os"

	"github.com/spf13/afero"
)

type AferoSink struct {
	fs       afero.Fs
	filePath string
}

func NewAferoSink(filePath string) (*AferoSink, error) {
	if len(filePath) == 0 {
		return nil, errors.New("empty filePath provided")
	}

	return &AferoSink{
		fs:       afero.NewOsFs(),
		filePath: filePath,
	}, nil
}

func (s *AferoSink) Read() ([]byte, error) {
	return afero.ReadFile(s.fs, s.filePath)
}

func (s *AferoSink) CanRead() error {
	_, err := afero.Exists(s.fs, s.filePath)
	return err
}

func (s *AferoSink) Write(signedData string) error {
	return afero.WriteFile(s.fs, s.filePath, []byte(signedData), 0640) // #nosec: G306
}

func (s *AferoSink) CanWrite() error {
	exists, _ := afero.Exists(s.fs, s.filePath)
	if !exists {
		return nil
	}

	file, err := s.fs.OpenFile(s.filePath, os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}
