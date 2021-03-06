package core

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Conf struct {
	AppName string `json:"appname"`
	Port    int    `json:"port"`
	AppRoot string `json:"approot"`
}

// RouteHandler callback func

type routerItem struct {
	method   string
	prefix   string
	handlers []gin.HandlerFunc
}

// App main
type App struct {
	Config       *Conf
	rootRouter   *gin.RouterGroup
	GinEngine    *gin.Engine
	routerGroups map[string]*gin.RouterGroup
	routers      map[string][]*routerItem
	midwares     map[string][]gin.HandlerFunc
}

// MidWare collect midwares
func (app *App) MidWare(
	group string,
	handler gin.HandlerFunc,
) {
	if _, ok := app.midwares[group]; !ok {
		app.midwares[group] = []gin.HandlerFunc{}
	}
	app.midwares[group] = append(app.midwares[group], handler)
}

type AppSet []*App

var Apps AppSet = []*App{}

// AllRouters apps all real router
func (apps *AppSet) AllRouters(index int) []string {
	app := (*apps)[index]
	ret := []string{}
	for group, routers := range app.routers {
		for _, router := range routers {
			currentRouter := group + router.prefix
			currentRouter = strings.Replace(currentRouter, "//", "/", -1)
			ret = append(ret, currentRouter)
		}
	}
	return ret
}

// Init init vars
func (app *App) Init() {
	if app.GinEngine == nil { // maybe assign outer
		app.GinEngine = gin.Default()
	}

	Apps = append(Apps, app)

	app.rootRouter = app.GinEngine.Group("/")
	app.routerGroups = map[string]*gin.RouterGroup{}
	app.routers = map[string][]*routerItem{}
	app.midwares = map[string][]gin.HandlerFunc{}
}

// Router collect and then make a router item
func (app *App) Router(
	group string,
	method string, // http verb list
	prefix string, // router
	handlers ...gin.HandlerFunc,
) {
	item := routerItem{
		method:   method,
		prefix:   prefix,
		handlers: handlers,
	}
	if _, ok := app.routers[group]; !ok {
		app.routers[group] = []*routerItem{}
	}
	app.routers[group] = append(app.routers[group], &item)

}

// SortedRouters asc sort by group length
func (app *App) SortedRouters() []string {
	_routers := [200][]string{}
	for group := range app.routers { // base on routers
		groupSet := _routers[len(group)]
		if groupSet == nil {
			groupSet = []string{}
		}
		groupSet = append(groupSet, group)
		_routers[len(group)] = groupSet
	}
	routers := []string{}
	for _, router := range _routers {
		if len(router) > 0 {
			routers = append(routers, router...)
		}
	}
	return routers
}

// AutoGroup make group logic match
func (app *App) AutoGroup(group string) *gin.RouterGroup {

	if routerGroup, ok := app.routerGroups[group]; ok {
		return routerGroup
	}

	groupStrs := strings.Split(group, "/")
	currentRouterGroup := app.rootRouter
	for _, currentGroupStr := range groupStrs {
		currentGroupStr = "/" + currentGroupStr
		if currentGroupStr == "/" {
			continue
		}
		if _, ok := app.routerGroups[currentGroupStr]; !ok {
			app.routerGroups[currentGroupStr] = currentRouterGroup.Group(currentGroupStr)
		}
		currentRouterGroup = app.routerGroups[currentGroupStr]
	}
	app.routerGroups[group] = currentRouterGroup
	return currentRouterGroup
}

func (app *App) Prepare() {
	sortedGroups := app.SortedRouters()
	for _, group := range sortedGroups {
		engine := app.AutoGroup(group)
		for _, midware := range app.midwares[group] {
			if engine == app.rootRouter {
				app.GinEngine.Use(midware) // root should effect on other routers not register
			}
			engine.Use(midware)
		}
		for _, router := range app.routers[group] {
			verbs := parseHTTPVerbs(router.method)
			ginHanlders := []gin.HandlerFunc{}
			ginHanlders = router.handlers
			for _, method := range verbs {
				if method == "GET" {
					if engine == app.rootRouter { // dont use root
						app.GinEngine.GET(
							router.prefix,
							ginHanlders...,
						)
					} else {
						engine.GET(
							router.prefix,
							ginHanlders...,
						)
					}

				} else if method == "POST" {
					if engine == app.rootRouter { // dont use root
						app.GinEngine.POST(
							router.prefix,
							ginHanlders...,
						)
					} else {
						engine.POST(
							router.prefix,
							ginHanlders...,
						)
					}
				} else if method == "PUT" {
					if engine == app.rootRouter { // dont use root
						app.GinEngine.PUT(
							router.prefix,
							ginHanlders...,
						)
					} else {
						engine.PUT(
							router.prefix,
							ginHanlders...,
						)
					}

				} else if method == "DELETE" {
					if engine == app.rootRouter { // dont use root
						app.GinEngine.DELETE(
							router.prefix,
							ginHanlders...,
						)
					} else {
						engine.DELETE(
							router.prefix,
							ginHanlders...,
						)
					}
				} else {
					panic(wrongMethodError{})
				}
			}
		}
	}

	if app.Config.Port == 0 {
		panic(noAddressError{})
	}
}

// Start start app
func (app *App) Start() {
	app.Prepare()
	app.GinEngine.Run(":" + strconv.Itoa(app.Config.Port))
}

func parseHTTPVerbs(method string) []string {
	methods := strings.Split(method, ",")
	return methods
}
