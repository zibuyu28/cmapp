package file

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// CopyFile copy a file
func CopyFile(sourceFile string, destinationFile string) error {
	// check sourceFile exists
	if _, err := os.Stat(sourceFile); err != nil {
		return err
	}

	// if exists then delete
	if _, err := os.Stat(destinationFile); err == nil {
		os.Remove(destinationFile)
	}

	path := destinationFile[:strings.LastIndex(destinationFile, "/")+1]
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}

	data, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destinationFile, data, 0777)
}

// TarGzDir compression dir to tar.gz
// 函数定义: 将src压缩到dstTar，运行异常时，返回错误信息
// 参数说明:
// 	src: 源文件/源文件夹
// 	dstTar: 结果文件(tar.gz格式)
// 	cover: 结果文件已存在，是否覆盖
// 	files: 当src为文件夹时，可从其目录下选择几个文件进行压缩；当src为文件时，无效
func TarGzDir(src, dstTar string, cover bool, files ...string) (err error) {
	if b, _ := PathExists(src); !b {
		return fmt.Errorf("src path is not exist: %s", src)
	}

	// 已存在 且 不覆盖
	if b, _ := PathExists(dstTar); b && !cover {
		fmt.Printf("destination dir [%s] exist\n", dstTar)
		return nil
	}

	fw, err := os.Create(dstTar)
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer func() {
		if er := tw.Close(); er != nil {
			err = er
		}
	}()

	fi, er := os.Stat(src)
	if er != nil {
		return er
	}

	srcBase, srcRelative := path.Split(path.Clean(src))
	if fi.IsDir() {
		if len(files) > 0 {
			for _, f := range files {
				ffi, err := os.Stat(path.Join(src, f))
				if err != nil {
					return err
				}
				tarFile(srcBase+srcRelative, f, tw, ffi)
			}
		} else {
			tarDir(srcBase, srcRelative, tw, fi)
		}
	} else {
		tarFile(srcBase, srcRelative, tw, fi)
	}

	return nil
}

// 因为要执行遍历操作，所以要单独创建一个函数
func tarDir(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := path.Join(srcBase, srcRelative)

	// 在结尾添加 "/"
	last := len(srcRelative) - 1
	if srcRelative[last] != os.PathSeparator {
		srcRelative += string(os.PathSeparator)
	}

	// 获取 srcFull 下的文件或子目录列表
	fis, er := ioutil.ReadDir(srcFull)
	if er != nil {
		return er
	}

	// 开始遍历
	for _, fi := range fis {
		if fi.IsDir() {
			tarDir(srcBase, srcRelative+fi.Name(), tw, fi)
		} else {
			tarFile(srcBase, srcRelative+fi.Name(), tw, fi)
		}
	}

	// 写入目录信息
	if len(srcRelative) > 0 {
		hdr, er := tar.FileInfoHeader(fi, "")
		if er != nil {
			return er
		}
		hdr.Name = srcRelative

		if er = tw.WriteHeader(hdr); er != nil {
			return er
		}
	}

	return nil
}

// 因为要在 defer 中关闭文件，所以要单独创建一个函数
func tarFile(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := path.Join(srcBase, srcRelative)

	// 写入文件信息
	hdr, er := tar.FileInfoHeader(fi, "")
	if er != nil {
		return er
	}
	hdr.Name = srcRelative

	if er = tw.WriteHeader(hdr); er != nil {
		return er
	}

	// 打开要打包的文件，准备读取
	fr, er := os.Open(srcFull)
	if er != nil {
		return er
	}
	defer fr.Close()

	// 将文件数据写入 tw 中
	if _, er = io.Copy(tw, fr); er != nil {
		return er
	}
	return nil
}

// PathExists returnutil.PathExists() true if path exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
