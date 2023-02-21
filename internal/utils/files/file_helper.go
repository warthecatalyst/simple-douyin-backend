package files

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"
)

// PathExists 判断是否存在路径
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(os.ErrNotExist, err) {
		return false, nil
	}
	return false, err
}

// ExtractCoverFromVideo 从视频中截取图像的第一帧
func ExtractCoverFromVideo(pathVideo, pathImg string) error {
	binPath := "./third_party/ffmpeg/"
	if runtime.GOOS == "windows" {
		binPath += "windows/"
	} else if runtime.GOOS == "darwin" {
		binPath += "darwin/"
	} else {
		binPath += "linux/"
	}

	frameExtractionTime := "0"
	image_mode := "image2"
	vtime := "0.001"

	// create the command
	cmd := exec.Command(binPath+"ffmpeg",
		"-i", pathVideo,
		"-y",
		"-f", image_mode,
		"-ss", frameExtractionTime,
		"-t", vtime,
		"-y", pathImg)

	// run the command and don't wait for it to finish. waiting exec is run
	// fmt.Println(cmd.String())
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

// SaveFileToLocal 把文件保存至本地
func SaveFileToLocal(savePath string, data *multipart.FileHeader) (string, error) {
	if exists, _ := PathExists(savePath); !exists {
		err := os.Mkdir(savePath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	src, err := data.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	timeLog := time.Now().Unix()
	fileName := GetFileNameWithoutExt(data.Filename)
	fileName += strconv.FormatInt(timeLog, 10) + path.Ext(data.Filename)
	out, err := os.Create(savePath + "/" + fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return fileName, err
}

func SaveDataToLocal(savePath string, data *[]byte, filename string) (string, error) {
	if exists, _ := PathExists(savePath); !exists {
		err := os.Mkdir(savePath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	src := bytes.NewReader(*data)
	timeLog := time.Now().Unix()
	fileName := GetFileNameWithoutExt(filename)
	fileName += strconv.FormatInt(timeLog, 10) + path.Ext(filename)
	out, err := os.Create(savePath + "/" + fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return fileName, err
}
