package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//
// File directory scanner implementation.
// @author rnojiri
//

// Scanner - Scans a directory and find files using regular expression
type Scanner struct {
	fileRegexp   *regexp.Regexp
	minFileSize  int64
	logger       *zap.Logger
	filesFound   []*File
	ignoredFiles []*File
	errorsFound  []error
}

// NewScanner - builds a new Scanner
func NewScanner(fileRegexp string, minFileSize int64, logger *zap.Logger) *Scanner {

	lf := []zapcore.Field{
		zap.String("package", "files"),
		zap.String("func", "NewScanner"),
	}

	logger.Info(fmt.Sprintf("creating a new scanner using regexp '%s' and minimum file size '%d'", fileRegexp, minFileSize), lf...)

	return &Scanner{
		fileRegexp:  regexp.MustCompile(fileRegexp),
		minFileSize: minFileSize,
		logger:      logger,
	}
}

var visitLogFields = []zapcore.Field{
	zap.String("package", "files"),
	zap.String("func", "visit"),
}

// visit - visits a file or a directory
func (s *Scanner) visit(path string, f os.FileInfo, err error) error {

	if err != nil {
		s.errorsFound = append(s.errorsFound, err)
		return err
	}

	if f.IsDir() {
		s.logger.Debug("ignoring directory:"+path, visitLogFields...)
		return nil
	}

	file := File{
		Name:             f.Name(),
		Path:             path,
		Size:             f.Size(),
		LastModification: f.ModTime(),
	}

	if s.fileRegexp.MatchString(path) {
		s.logger.Debug("file name matches with regexp: "+path, visitLogFields...)

		if file.Size < s.minFileSize {
			s.logger.Debug(fmt.Sprintf("file does not have the minimum size: %s (%d/%d)", path, file.Size, s.minFileSize), visitLogFields...)
			file.Ignored = true
			s.ignoredFiles = append(s.ignoredFiles, &file)
			return nil
		}

		file.Ignored = false
		s.filesFound = append(s.filesFound, &file)
		return nil
	}

	file.Ignored = true
	s.ignoredFiles = append(s.ignoredFiles, &file)
	return nil
}

// Scan - searches recursively for files with the same regexp pattern selected
func (s *Scanner) Scan(directoryRoot string) (*ScanResult, error) {

	s.filesFound = []*File{}
	s.ignoredFiles = []*File{}
	s.errorsFound = []error{}

	err := filepath.Walk(directoryRoot, s.visit)

	return &ScanResult{
		Files:   s.filesFound,
		Ignored: s.ignoredFiles,
		Errors:  s.errorsFound,
	}, err
}
