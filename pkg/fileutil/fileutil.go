package fileutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	CmdPath = ""
)

// 文件是否存在
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 文件夹是否为空
func DirIsEmpty(dirname string) bool {
	dir, _ := ioutil.ReadDir(dirname)
	return len(dir) == 0
}

// basePath是固定目录路径
func CreateDir(basePath, folderName string) (string, error) {
	folderPath := filepath.Join(basePath, folderName)
	var err error
	if _, err = os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步
		// 先创建文件夹
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			return folderPath, err
		}
		// 再修改权限
		err = os.Chmod(folderPath, 0755)
		if err != nil {
			return folderPath, err
		}
	}
	return folderPath, nil
}

func RemoveDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func IsUrlMovie(s string) bool {
	//解析url
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		//网络地址错误
		fmt.Println("网络地址错误")
		return false
	}
	extname := path.Ext(uri.Path)
	return extname == ".mp4" || extname == ".rmvb"
}

func Md5File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 列举目录内所有文件
// prefix 调用时传""
func ListFileNames(path string, prefix string) ([]string, error) {
	result := make([]string, 0)
	//以只读的方式打开目录
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return result, nil
	}
	//延迟关闭目录
	defer f.Close()
	fileInfo, _ := f.Readdir(-1)
	for _, info := range fileInfo {
		if info.IsDir() { //文件夹里只有一个目录
			temp, err := ListFileNames(filepath.Join(path, info.Name()), filepath.Join(prefix, info.Name()))
			if err != nil {
				return result, err
			}
			result = append(result, temp...)
		} else {
			result = append(result, filepath.Join(prefix, info.Name()))
		}
	}
	return result, nil
}

// 切割大文件为多个小文件, 返回切片后的路径数组, 如果文件大小小于chunkSize, 则返回空数组
// saveDir 保存路径
// chunkSize切割大小,1M 则1024*1024
func SplitFile(filePth string, saveDir string, chunkSize int64) ([]string, error) {
	ret := make([]string, 0)
	fileInfo, err := os.Stat(filePth)
	if err != nil {
		//文件不存在
		return nil, err
	}
	if fileInfo.Size() <= chunkSize {
		//如果文件大小小于chunkSize, 则返回空数组
		return ret, nil
	}

	//检测保存的文件夹是否存在, 如果不存在，则创建
	if _, err = os.Stat(saveDir); os.IsNotExist(err) {
		// 必须分成两步
		// 先创建文件夹
		err = os.MkdirAll(saveDir, 0755)
		if err != nil {
			return nil, err
		}
		// 再修改权限
		err = os.Chmod(saveDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	//切片个数
	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))
	//后缀名
	fileExt := filepath.Ext(filePth)
	fileNameOnly := strings.TrimSuffix(filepath.Base(filePth), fileExt)

	fi, err := os.OpenFile(filePth, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	b := make([]byte, chunkSize)
	var i int64 = 1
	for ; i <= int64(num); i++ {
		fi.Seek((i-1)*(chunkSize), 0)
		if len(b) > int((fileInfo.Size() - (i-1)*chunkSize)) {
			b = make([]byte, fileInfo.Size()-(i-1)*chunkSize)
		}
		fi.Read(b)
		//写入新文件
		newFileName := fmt.Sprintf("%s_%d%s", fileNameOnly, i, fileExt)
		newFilePth := filepath.Join(saveDir, newFileName)
		f, err := os.OpenFile(newFilePth, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return nil, err
		}
		f.Write(b)
		f.Close()
		ret = append(ret, newFilePth)
	}
	return ret, nil
}

func ReadFile(filePath string) ([]byte, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

// 覆盖写文件
func WriteFile(content []byte, filePath string) error {
	//写入文件
	return ioutil.WriteFile(filePath, content, 0666)
}

// 拷贝文件
func CopyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

// 从指定的目录里查找资源, 查找的目录包括$GOPATH, $GOPATH/src/resources, $GOPATH/resources, appPath, workPath/docs, /opt/goproject等
func FindResourcePth(fileName string, pathsarr ...string) string {
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	workPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	paths := make([]string, 0)
	if len(pathsarr) > 0 {
		for _, pth := range pathsarr {
			paths = append(paths, filepath.Join(pth, fileName))
		}
	}
	paths = append(paths, filepath.Join(appPath, fileName))
	paths = append(paths, filepath.Join(workPath, fileName))
	if CmdPath != "" {
		paths = append(paths, filepath.Join(CmdPath, fileName))
	}
	//根据环境变量查找路径
	gopth := os.Getenv("GOPATH")
	if gopth != "" {
		paths = append(paths, filepath.Join(gopth, fileName))
		paths = append(paths, filepath.Join(gopth, "resources", fileName))
		paths = append(paths, filepath.Join(gopth, "src", "resources", fileName))
	} else {
		fmt.Println("***************************获取环境变量失败，如果是使用supervisor管理应用，请在supervisord.conf [supervisord]增加environment = GOPATH=\"/opt/goproject\" ;环境变量 *******************************")
	}
	paths = append(paths, filepath.Join(workPath, "docs", fileName))
	paths = append(paths, filepath.Join("/srv/goproject", fileName))
	paths = append(paths, filepath.Join("/data/zyvolumes/resources", fileName))
	for i := 0; i < len(paths); i++ {
		if FileExists(paths[i]) {
			fmt.Println("从paths: ", paths[i], " 中已找到资源:", fileName)
			return paths[i]
		}
	}
	fmt.Println("从paths: ", paths, " 中未找到资源:", fileName)
	return ""
}
