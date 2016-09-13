package main

import "github.com/russross/blackfriday"
import "github.com/hoisie/mustache"
import (
    "os"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "strings"
    "errors"
    "time"
    "sync"
)

const (
    commonHTMLFlags = 0 |
        blackfriday.HTML_USE_XHTML |
        blackfriday.HTML_USE_SMARTYPANTS |
        blackfriday.HTML_SMARTYPANTS_FRACTIONS |
        blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
        blackfriday.HTML_TOC

    commonExtensions = 0 |
        blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
        blackfriday.EXTENSION_TABLES |
        blackfriday.EXTENSION_FENCED_CODE |
        blackfriday.EXTENSION_AUTOLINK |
        blackfriday.EXTENSION_STRIKETHROUGH |
        blackfriday.EXTENSION_SPACE_HEADERS |
        blackfriday.EXTENSION_HEADER_IDS |
        blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
        blackfriday.EXTENSION_DEFINITION_LISTS
)

func main() {
    var dirPath string
    if len(os.Args) < 2 {
        var dirErr error
        dirPath,dirErr = os.Getwd()
        if dirErr != nil {
            showError( dirErr.Error() )
        }
    } else {
        dirPath = os.Args[1]
    }

    fileInfo, err := os.Stat(dirPath)
    if err != nil {
        switch  {
        case os.IsNotExist(err):
            showError("文件不存在:" + dirPath)
        case os.IsPermission(err):
            showError("权限不足无法操作文件" + dirPath)
        case !fileInfo.Mode().IsDir():
            showError("文件不是一个文件夹" + dirPath)
        }
    }

    delOldHtmlFile(dirPath)

    fileList,globErr := filepath.Glob(dirPath + string(os.PathSeparator) + "*.md")
    if globErr != nil {
        showError("提取目录" + dirPath + "下md文件失败:" + globErr.Error())
    }
    group := sync.WaitGroup{}
    group.Add(len(fileList))
    for _,file := range fileList{
        go func(file string) {
            renderErr := convert(file)
            if renderErr != nil {
                fmt.Println("转换文件失败:"+file+" 错误信息:" + renderErr.Error())
            } else {
                fmt.Println("转换文件成功:"+file)
            }
            group.Done()
        }(file)
    }
    group.Wait()
}

// 转化md文件作为html文件
func convert(file string) error {

    fileInfo, err := os.Stat(file)
    if err != nil {
        switch  {
        case os.IsNotExist(err):
            return errors.New("文件不存在:" + file)
        case os.IsPermission(err):
            return errors.New("权限不足无法操作文件" + file)
        case !fileInfo.Mode().IsRegular():
            return errors.New("文件不是一个普通文件" + file)
        }
    }

    input, errRead := ioutil.ReadFile(file)
    if errRead != nil {
        return errors.New("读取文件" + file + "失败:" + errRead.Error())
    }

    render := blackfriday.HtmlRenderer(commonHTMLFlags, "", "")
    body := blackfriday.MarkdownOptions(input, render, blackfriday.Options{
        Extensions: commonExtensions,
    })

    m := map[string]interface{}{
        "title":  strings.TrimSuffix(fileInfo.Name(), ".md"),
        "body": string(body),
    }

    templateContent,readTemplateErr := Asset("data/md.template")
    if readTemplateErr != nil {
        return errors.New("读取模板文件失败:" + readTemplateErr.Error())
    }

    template, templateErr := mustache.ParseString(string(templateContent))
    if templateErr != nil {
        return errors.New("解析模板失败" + templateErr.Error())
    }

    output := template.Render(m)
    outputFile := strings.TrimSuffix(file, ".md") + "_" + time.Now().Format("2006-01-02_150405") + ".html";
    errWrite := ioutil.WriteFile(outputFile, []byte(output), 0644);
    if errWrite != nil {
        return errors.New("写入文件" + outputFile + "失败:" + errWrite.Error())
    }

    return nil
}

// 删除旧html
func delOldHtmlFile(dirPath string) {
    fileList,globErr := filepath.Glob(dirPath + string(os.PathSeparator) + "*.html")
    if globErr != nil {
        showError("获取目录" + dirPath + "下旧html文件失败")
    }

    for _,file :=range fileList {
        if err := os.Remove(file); err != nil {
            fmt.Println("删除旧文件" + file + "失败")
        }
    }

}

// 显示错误
func showError(msg string)  {
    fmt.Println(msg)
    os.Exit(1)
}
