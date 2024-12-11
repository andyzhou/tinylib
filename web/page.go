package web

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"html/template"
	"os"
	"sync"
)

/*
 * tpl page face
 * - support static html page generate
 */

//page config
type PageConfig struct {
	TplPath string
	StaticPath string
}

//face info
type Page struct {
	cfg           *PageConfig
	extFuncMap    map[string]interface{}
	shareTplFiles []string
	sync.RWMutex
}

//construct
func NewPage() *Page {
	this := &Page{
		cfg: &PageConfig{
			TplPath: ".",
			StaticPath: ".",
		},
		extFuncMap: map[string]interface{}{},
		shareTplFiles: []string{},
	}
	return this
}

//set config, STEP-1
func (f *Page) SetConfig(cfg *PageConfig) error {
	//check
	if cfg == nil {
		return errors.New("invalid parameter")
	}
	f.cfg = cfg
	return nil
}

//parse tpl, STEP-2
//return template, error
func (f *Page) ParseTpl(
		mainTplFile string,
		tplPaths ...string,
	) (*template.Template, error) {
	var (
		tplPath string
		err error
	)
	//check
	if mainTplFile == "" {
		return nil, errors.New("invalid parameter")
	}
	if tplPaths != nil && len(tplPaths) > 0 {
		tplPath = tplPaths[0]
	}

	//init new template obj
	tpl := template.New(mainTplFile)

	//format tpl full path
	mainTplFullPath := fmt.Sprintf("%v/%v", f.cfg.TplPath, mainTplFile)
	if tplPath != "" {
		mainTplFullPath = fmt.Sprintf("%v/%v", tplPath, mainTplFile)
	}

	//setup final tpl files
	finalTplFiles := make([]string, 0)
	if f.shareTplFiles != nil && len(f.shareTplFiles) > 0 {
		finalTplFiles = append(finalTplFiles, f.shareTplFiles...)
	}
	finalTplFiles = append(finalTplFiles, mainTplFullPath)

	//begin parse tpl files
	tpl, err = tpl.ParseFiles(finalTplFiles...)
	return tpl, err
}

//get tpl content, STEP-3 [option]
func (f *Page) GetTplContent(
		tpl *template.Template,
		tplData map[string]interface{},
	) (string, error) {
	//check
	if tpl == nil {
		return "", errors.New("invalid parameter")
	}

	//get tpl content
	tplBuff := bytes.NewBuffer(nil)
	err := tpl.Execute(tplBuff, tplData)
	if err != nil {
		return "", err
	}

	//must unescape html string for show
	originTpl := tplBuff.String()
	originTpl = html.UnescapeString(originTpl)
	return originTpl, nil
}

//gen html page file, STEP-3 [option]
func (f *Page) GenHtmlPage(
	tpl *template.Template,
	pageFilePath string) error {
	//check
	if tpl == nil || pageFilePath == "" {
		return errors.New("invalid parameter")
	}

	//create page
	pageFileFullPath := fmt.Sprintf("%v/%v", f.cfg.StaticPath, pageFilePath)
	out, err := os.Create(pageFileFullPath)
	if err != nil {
		return err
	}

	//output page file
	defer out.Close()
	err = tpl.Execute(out, pageFileFullPath)
	return err
}

//add ext func
func (f *Page) AddExtFunc(name string, fun interface{}) error {
	//check
	if name == "" || fun == nil {
		return errors.New("invalid parameter")
	}

	//check run env
	f.Lock()
	defer f.Unlock()
	_, ok := f.extFuncMap[name]
	if ok {
		return errors.New("ext func name is exists")
	}

	//save into run env
	f.extFuncMap[name] = fun
	return nil
}

//add shared tpl files
func (f *Page) AddShareTplFiles(fileNames ...string) error {
	//check
	if fileNames == nil || len(fileNames) <= 0 {
		return errors.New("invalid parameter")
	}

	//fill into run env
	for _, fileName := range fileNames {
		fileFullPath := fmt.Sprintf("%v/%v", f.cfg.TplPath, fileName)
		f.shareTplFiles = append(f.shareTplFiles, fileFullPath)
	}
	return nil
}

//reset shared tpl files
func (f *Page) ResetSharedTplFiles() {
	f.shareTplFiles = []string{}
}

//check or create dir
func (f *Page) CheckOrCreateOneDir(dirPath string) error {
	//check
	if dirPath == "" {
		return errors.New("invalid parameter")
	}

	//detect dir path
	fullDirPath := fmt.Sprintf("%v/%v", f.cfg.StaticPath, dirPath)
	_, err := os.Stat(fullDirPath)
	if err == nil {
		return err
	}
	bRet := os.IsExist(err)
	if bRet {
		return nil
	}

	//create dir path
	err = os.Mkdir(fullDirPath, 0777)
	return err
}