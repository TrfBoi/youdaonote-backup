package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
)

const (
	CSTKCookieKey = "YNOTE_CSTK"

	getRootURL = "https://note.youdao.com/yws/api/personal/file?method=getByPath&cstk=%s&keyfrom=web" // 获取指定path的信息, post
	// 获取指定目录下的文件及目录信息, get。url后面的len参数需要设置, 不然可能出现大文件夹缺失的逻辑
	listURL = "https://note.youdao.com/yws/api/personal/file/%s?all=true&method=listPageByParentId&cstk=%s&len=%d"
	noteURL = "https://note.youdao.com/yws/api/personal/sync?method=download&sev=j1&keyfrom=web" // 下载笔记信息地址, post
	// 笔记中图片与附件地址
	// pictureURL = "https://note.youdao.com/yws/res/%d/%s"

	// 相对于BackupLocalPath的图片存放路径
	picturePath = "picture"
)

var (
	CSTK            string
	BackupLocalPath string
	MaxDirFileCount int
	// 存放目录路径 key dirID, value dirPath
	path = map[string]string{}

	pictureRE *regexp.Regexp
	attachRE  *regexp.Regexp
)

// YouDaoNoteDir  有道云笔记目录信息表示
type YouDaoNoteDir struct {
	Entries []*YouDaoNoteFile `json:"entries"`
	Count   int               `json:"count"`
}

// YouDaoNoteFile 有道云笔记文件信息表示
type YouDaoNoteFile struct {
	FileEntry struct {
		UserId         string `json:"userId"`
		Id             string `json:"id"`
		Version        int    `json:"version"`
		Name           string `json:"name"`
		ParentId       string `json:"parentId"`
		FileNum        int    `json:"fileNum"`
		DirNum         int    `json:"dirNum"`
		SubTreeFileNum int    `json:"subTreeFileNum"`
		SubTreeDirNum  int    `json:"subTreeDirNum"`
		EntryType      int    `json:"entryType"`
		Dir            bool   `json:"dir"`
		OrgEditorType  int    `json:"orgEditorType"`
	} `json:"fileEntry"`
	FileMeta struct {
		Title       string `json:"title"`
		ContentType string `json:"contentType"`
	} `json:"fileMeta"`
}

func Init() {
	CSTK = Cookies[CSTKCookieKey].Value
	if len(CSTK) == 0 {
		panic("cookie中没有YNOTE_CSTK, 请检查cookie")
	}

	if err := os.MkdirAll(BackupLocalPath, 0755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Join(BackupLocalPath, picturePath), 0755); err != nil {
		panic(err)
	}

	var err error
	if pictureRE, err = regexp.Compile("src=\"(https://note\\.youdao\\.com/yws/res/\\d+/(\\w+))\""); err != nil {
		panic(err)
	}
	if attachRE, err = regexp.Compile("path=\"(https://note\\.youdao\\.com/yws/res/\\d+/\\w+)\".*?filename=\"(.*?)\""); err != nil {
		panic(err)
	}
}

// BackupAllNote 备份所有笔记
func BackupAllNote() {
	// 1. 获取根目录文件夹 id
	rootDir := getRootDir()
	// 2. 递归遍历下载
	downloadDir(rootDir)
}

func getRootDir() *YouDaoNoteFile {
	data, err := PostForm(fmt.Sprintf(getRootURL, CSTK), map[string]string{
		"path":   "/",
		"entire": "true",
		"purge":  "false", // 经过验证可选
		"cstk":   CSTK,    // 经过验证可选
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("获取根目录ID出错, cookie可能出错或程序已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新")
	}
	rootFile := &YouDaoNoteFile{}
	if err = json.Unmarshal(data, rootFile); err != nil {
		fmt.Println(err)
		fmt.Println("根目录结果反序列化失败, 程序已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新")
	}
	return rootFile
}

func downloadDir(d *YouDaoNoteFile) {
	if d.FileEntry.Name == "ROOT" {
		path[d.FileEntry.Id] = filepath.Join(BackupLocalPath)
	} else {
		path[d.FileEntry.Id] = filepath.Join(path[d.FileEntry.ParentId], d.FileEntry.Name)
	}
	fileNum := d.FileEntry.FileNum + d.FileEntry.DirNum
	if fileNum > MaxDirFileCount {
		MaxDirFileCount = fileNum
	}
	data, err := Get(fmt.Sprintf(listURL, d.FileEntry.Id, CSTK, MaxDirFileCount))
	if err != nil {
		fmt.Println(err)
		fmt.Println("获取目录结构失败, 程序已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新")
	}
	dir := &YouDaoNoteDir{}
	if err = json.Unmarshal(data, dir); err != nil {
		fmt.Println(err)
		fmt.Println("目录结构反序列化失败, 程序已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新")
	}
	for _, file := range dir.Entries {
		if file.FileEntry.Dir {
			downloadDir(file)
		} else {
			downloadFile(file)
		}
	}
}

func downloadFile(file *YouDaoNoteFile) {
	fmt.Println(fmt.Sprintf("备份 %s ", file.FileEntry.Name))
	data, err := PostForm(noteURL, map[string]string{
		"fileId":     file.FileEntry.Id,
		"version":    "-1",
		"convert":    "true", // 需要为true, 不然图片引用非url, 后需代码无法处理
		"editorType": "0",    // 经过测试与返回格式有关, 不能设置为1, 会返回xml格式
		"cstk":       CSTK,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println(fmt.Sprintf("下载 %s 失败,程序可能已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新", file.FileEntry.Name))
		return
	}
	// 下载图片与附件以及替换文本, 让本地可见图片
	handleLink(data)
	data = replacePictureLink(data, file)
	if err := os.MkdirAll(path[file.FileEntry.ParentId], 0755); err != nil {
		log.Fatalln(fmt.Sprintf("创建 %s 目录失败", path[file.FileEntry.ParentId]))
	}
	file.FileEntry.Name = strings.ReplaceAll(file.FileEntry.Name, ".note", ".html")
	fd, err := os.Create(filepath.Join(path[file.FileEntry.ParentId], file.FileEntry.Name))
	if err != nil {
		log.Fatalln("传入的本地目录有误, 请检查")
	}
	defer fd.Close()
	if _, err := fd.Write(data); err != nil {
		fmt.Println(err)
	}
}

func handleLink(data []byte) {
	dataStr := string(data)
	handlePicture(&dataStr)
	handleAttachFile(&dataStr)
}

func handlePicture(dataStr *string) {
	links := pictureRE.FindAllStringSubmatch(*dataStr, -1)
	for _, link := range links {
		rsp, err := get(link[1])
		if err != nil {
			fmt.Println(err)
			fmt.Println(fmt.Sprintf("下载 %s 失败, 程序可能已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新", link))
			continue
		}
		picture, _, err := image.Decode(rsp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		rsp.Close()
		fd, err := os.Create(filepath.Join(BackupLocalPath, picturePath, link[2]+".png"))
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := png.Encode(fd, picture); err != nil {
			fmt.Println(err)
			continue
		}
		fd.Close()
	}
}

func handleAttachFile(dataStr *string) {
	links := attachRE.FindAllStringSubmatch(*dataStr, -1)
	for _, link := range links {
		rsp, err := Get(link[1])
		if err != nil {
			fmt.Println(err)
			fmt.Println(fmt.Sprintf("下载 %s 失败, 程序可能已经失效请联系作者(bioittang@163.com), 作者有时间将尝试更新", link))
			continue
		}
		fd, err := os.Create(filepath.Join(BackupLocalPath, picturePath, link[2]))
		if err != nil {
			log.Fatalln("传入的本地目录有误, 请检查")
		}
		if _, err := fd.Write(rsp); err != nil {
			fmt.Println(err)
		}
		fd.Close()
	}
}

func replacePictureLink(data []byte, file *YouDaoNoteFile) []byte {
	dataStr := string(data)
	links := pictureRE.FindAllStringSubmatch(dataStr, -1)
	for _, link := range links {
		dataStr = strings.ReplaceAll(dataStr, link[1], filepath.Join(getPictureRP(file), link[2]+".png"))
	}
	return []byte(dataStr)
}

func getPictureRP(file *YouDaoNoteFile) string {
	fileDir := filepath.Join(path[file.FileEntry.ParentId])
	tmpPath := strings.TrimPrefix(fileDir, filepath.Join(BackupLocalPath))
	// 为了适配不同系统, 路径分割符不能是'/'
	n := strings.Count(tmpPath, string(os.PathSeparator))
	var strBuilder strings.Builder
	strBuilder.Grow(n * 3)
	for i := 0; i < n; i++ {
		strBuilder.WriteString("..")
		strBuilder.WriteString(string(os.PathSeparator))
	}
	return filepath.Join(strBuilder.String(), picturePath)
}
