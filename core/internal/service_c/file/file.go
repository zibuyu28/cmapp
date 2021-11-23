package file

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
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
	abs, _ := filepath.Abs(r.baseDir)
	_ = os.MkdirAll(abs, os.ModePerm)
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
	log.Debugf(context.Background(),"file name [%s]", fileName)
	s := md5.MD5(fileName)
	filePath := filepath.Join(r.baseDir, s)
	open, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "op file [%s]", filePath)
	}
	return open, nil
}