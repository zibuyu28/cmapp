package fabtool

import (
	"context"
	"fmt"
	"github.com/zibuyu28/cmapp/core/pkg/ag/base"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
)

type RT interface{}
type instanceTool func(ctx context.Context) RT

var mf = map[string]instanceTool{
	"configtxgen":   newConfigTXGen,
	"cryptogen":     newCryptoGen,
	"peer":          newPeer,
	//"configtxlator": newConfigTXLator,
}

func NewTool(ctx context.Context, driveruuid, toolName, version string) (RT, error) {
	pwd, _ := os.Getwd()
	toolFullName := filepath.Join(pwd, "drivers", driveruuid, fmt.Sprintf("tool/%s/%s", version, toolName))
	_, err := os.Stat(toolFullName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf(ctx, "%s not exist", toolName)
			_ = os.MkdirAll(filepath.Dir(toolFullName), os.ModePerm)
			err = downLoadTool(ctx, toolName, version, toolFullName)
			if err != nil {
				return nil, errors.Wrap(err, "download tool")
			}
		} else {
			return nil, errors.Wrapf(err, "check tool [%s] exist", toolFullName)
		}
	}

	err = os.Chmod(toolFullName, os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "chmod tool file mode")
		return nil, err
	}
	tool, ok := mf[toolName]
	if !ok {
		err = fmt.Errorf("tool(%s) not register", toolName)
		return nil, err
	}
	return tool(ctx), nil
}

// extractVersionInFormat extract version from string in format
func extractVersionInFormat(version string) (fv string, err error) {
	// xxxxx-v1.3.4
	// [\d]+\.[\d]+\.[\d]+
	compile := regexp.MustCompile("[\\d]+\\.[\\d]+\\.[\\d]+")
	findString := compile.FindString(version)
	if len(findString) == 0 {
		err = errors.Errorf("version(%s) not in format", version)
		return
	}
	fv = findString
	return
}

func downLoadTool(ctx context.Context, toolName, version string, toolFullName string) error {
	formatVersion, err := extractVersionInFormat(version)
	if err != nil {
		return errors.Wrap(err, "extract version")
	}
	log.Debugf(ctx, "get os [%s] and arch [%s]", runtime.GOOS, runtime.GOARCH)
	remoteFileName := fmt.Sprintf("fabric-%s-%s-%s-%s", toolName, runtime.GOOS, runtime.GOARCH, formatVersion)
	core := ag.Core{ApiVersion: base.V1}
	fc, err := core.DownloadFile(remoteFileName)
	if err != nil {
		return errors.Wrap(err, "download file from driver")
	}
	if len(fc) < 5000 {
		log.Debugf(ctx, "download result [%s]", string(fc))
		return errors.Errorf("download tool(%s) content not correct", toolFullName)
	}
	err = ioutil.WriteFile(toolFullName, fc, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("write content for tool : %s", toolFullName))
	}
	return nil
}
