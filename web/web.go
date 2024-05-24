package web

import "sync"

/*
 * web simple face
 * - base on gin engine
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//global variable
var (
	_web *Web
	_webOnce sync.Once
)

//face info
type Web struct {
	app *App
	cookie *Cookie
	page *Page
	Base
}

//get single instance
func GetWeb() *Web {
	_webOnce.Do(func() {
		_web = NewWeb()
	})
	return _web
}

//construct
func NewWeb() *Web {
	this := &Web{
		app: NewApp(),
		cookie: NewCookie(),
		page: NewPage(),
	}
	return this
}

//get sub face
func (f *Web) GetApp() *App {
	return f.app
}
func (f *Web) GetCookie() *Cookie {
	return f.cookie
}
func (f *Web) GetPage() *Page {
	return f.page
}