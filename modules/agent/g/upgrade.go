package g

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

func AgentSelfUpgrade() {
	for {
		select {
		case upgradeArgs := <-UpgradeChannel:
			log.Printf("AgentSelfUpgrade_channel:%+v", upgradeArgs)
			binUrl := fmt.Sprintf("%s/%s_%s", upgradeArgs.WgetUrl, "bin", upgradeArgs.Version)
			cfgUrl := fmt.Sprintf("%s/%s_%s", upgradeArgs.WgetUrl, "cfg", upgradeArgs.Version)
			agentPath := fmt.Sprintf("%s/falcon-agent", Config().AppBaseDir)
			cfgPath := fmt.Sprintf("%s/cfg.json", Config().AppBaseDir)
			var err error
			switch upgradeArgs.Type {
			case 0:
				err = downloadReplaceFile(agentPath, binUrl, upgradeArgs.BinFileMd5)
			case 1:
				err = downloadReplaceFile(cfgPath, cfgUrl, upgradeArgs.CfgFileMd5)
			case 2:
				errBin := downloadReplaceFile(agentPath, binUrl, upgradeArgs.BinFileMd5)
				errCfg := downloadReplaceFile(cfgPath, cfgUrl, upgradeArgs.CfgFileMd5)
				if errBin == nil && errCfg == nil {
					err = nil
				}
			}

			InUpgrading = false
			if err == nil {
				log.Printf("升级完成,重置升级状态")
				// 升级完成后获取当前agent的pid 然后给自己发送kill 信号
				pid := os.Getpid()
				thisPro, _ := os.FindProcess(pid)
				thisPro.Signal(os.Kill)
			}
			break
		}
	}
}

func BackUpFile(old string) (err error) {
	_, err = os.Stat(old)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File:%+v does not exist:%+v", old, err)
			return
		}
	}
	t := time.Now().Format("20060102")
	new := fmt.Sprintf("%s_%s_%s", old, "bak", t)
	mvCmd := fmt.Sprintf("/bin/mv -f %s %s", old, new)
	resStr := ExeSysCommand(mvCmd)
	if resStr == "FAILED" {
		return errors.New("MV failed")
	}
	return nil
}

func CopyFile(dstName, srcName string) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	defer dst.Close()
	_, errcopy := io.Copy(dst, src)
	if errcopy != nil {
		log.Printf("backUpFile_CopyFile_error:%+v", errcopy)
	}
	log.Printf("文件:%s 备份完成", srcName)

}

func downloadReplaceFile(filepath string, url string, argMd5 string) (err error) {
	log.Printf("downloadFile:%+v,%+v,", filepath, url)
	t := time.Now().Format("20060102")
	new := fmt.Sprintf("%s_%s_%s", filepath, "bak", t)
	cpCmd := fmt.Sprintf("/bin/cp -f %s %s", new, filepath)
	//先备份下
	err = BackUpFile(filepath)
	if err != nil {
		return
	}
	//开始下载
	out, err := os.Create(filepath)
	if err != nil {
		log.Printf("downloadFile_create_file_error:%+v", err)
		return err
	}
	defer out.Close()

	//get data from http
	resp, err := http.Get(url)
	//打上Range header 支持断点续传
	resp.Header.Set("Range", "bytes=0-")
	if err != nil {
		log.Printf("downloadFile_wget_error:%+v", err)
		//下载失败应该把文件copy 回去
		resStr := ExeSysCommand(cpCmd)
		if resStr == "FAILED" {
			return errors.New("CP failed")
		}
		err = os.Chmod(filepath, 0755)
		if err != nil {
			log.Printf("downloadFile_chmod_error:%+v", err)
			return err
		}
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		//下载失败应该把文件copy 回去
		log.Printf("downloadFile_wget_errcode:%+v", http.StatusOK)
		resStr := ExeSysCommand(cpCmd)
		if resStr == "FAILED" {
			return errors.New("CP failed")
		}
		err = os.Chmod(filepath, 0755)
		if err != nil {
			log.Printf("downloadFile_chmod_error:%+v", err)
			return err
		}
		err = os.Chmod(filepath, 0755)
		log.Printf("回滚copy_file 成功")

		return fmt.Errorf("bad status: %s", resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("downloadFile_io.Copy_error:%+v", err)
		//io.copy失败应该把文件copy 回去
		resStr := ExeSysCommand(cpCmd)
		if resStr == "FAILED" {
			return errors.New("CP failed")
		}
		err = os.Chmod(filepath, 0755)
		if err != nil {
			log.Printf("downloadFile_chmod_error:%+v", err)
			return err
		}
		return err
	}
	log.Printf("url文件:%s 下载完成", url)
	//对比MD5
	md5Same := CheckMd5(filepath, argMd5)
	if md5Same == false {
		log.Printf("md5不匹配,回滚")
		//md5 check失败应该把文件copy 回去
		resStr := ExeSysCommand(cpCmd)
		if resStr == "FAILED" {
			return errors.New("CP failed")
		}
		err = os.Chmod(filepath, 0755)
		if err != nil {
			log.Printf("downloadFile_chmod_error:%+v", err)
			return err
		}
		return errors.New("Md5 check Failed")
	}

	log.Printf("Md5 check OK")
	err = os.Chmod(filepath, 0755)
	if err != nil {
		log.Printf("downloadFile_chmod_error:%+v", err)
		return err
	}
	err = os.Chmod(filepath, 0755)
	if err != nil {
		log.Printf("downloadFile_chmod_error:%+v", err)
		return err
	}
	return nil
}

func CheckMd5(fp, argMd5 string) (same bool) {
	cmd := "/usr/bin/md5sum"
	command := fmt.Sprintf("%s %s", cmd, fp)
	res := ExeSysCommand(command)
	newMd5 := strings.Split(res, " ")[0]
	if argMd5 == newMd5 {
		same = true
	}
	return
}

func ExeSysCommand(cmdStr string) string {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	opBytes, err := cmd.Output()
	if err != nil {
		log.Debugf("ExeSysCommand:%s ,error:%+v", cmdStr, err)
		return "FAILED"
	}
	return string(opBytes)
}
