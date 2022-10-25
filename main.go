package main

import (
	"fmt"

	"github.com/spf13/pflag"
)

func main() {
	pflag.StringVarP(&RawCookies, "cookies", "c", "", "必填, 有道云笔记Cookies, 格式: key1=value1;key2=value2")
	pflag.StringVarP(&BackupLocalPath, "dir", "d", "./", "可选, 本地备份目录, 默认值: ./")
	pflag.IntVarP(&MaxDirFileCount, "max_dir_file_count", "m", 30, "可选, 有道云最大目录下的目录与文件个数")
	pflag.IntVarP(&Timeout, "http_timeout", "t", 60, "可选")
	pflag.Parse()

	if len(RawCookies) == 0 {
		panic("必须设置Cookies程序才可正常运行")
	}
	if err := ParseCookies(); err != nil {
		panic("设置Cookies格式有误, 请检查重试")
	}

	Init()
	fmt.Println("\n------------------- 开始备份 ------------------")
	BackupAllNote()
	fmt.Println("\n------------------- 备份结束 ------------------")
}
