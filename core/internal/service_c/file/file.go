package file

import (
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/md5"
	"os"
	"path/filepath"
)

// RFD remote file driver
type RFD struct {
	baseDir string
}

const fileDBDir = "./filedb"

var RFDi = RFD{baseDir: fileDBDir}


// UploadFile upload file
func (r *RFD) UploadFile(fileName string, saveExec func(string) error) error {
	s := md5.MD5(fileName)
	filePath := filepath.Join(r.baseDir, s)
	err := saveExec(filePath)
	if err != nil {
		return errors.Wrap(err, "save exec")
	}
	return nil
}

// DownloadFile download file
func (r *RFD) DownloadFile(fileName string) (*os.File,error) {
	s := md5.MD5(fileName)
	filePath := filepath.Join(r.baseDir, s)
	open, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "op file [%s]", filePath)
	}
	return open, nil
}