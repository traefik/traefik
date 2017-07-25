# Negroni [![GoDoc](https://godoc.org/github.com/urfave/negroni?status.svg)](http://godoc.org/github.com/urfave/negroni) [![wercker status](https://app.wercker.com/status/13688a4a94b82d84a0b8d038c4965b61/s "wercker status")](https://app.wercker.com/project/bykey/13688a4a94b82d84a0b8d038c4965b61)

在Go语言里，Negroni 是一个很地道的 web 中间件，它是微型，非嵌入式，并鼓励使用原生 `net/http` 处理器的库。

如果你用过并喜欢 [Martini](http://github.com/go-martini/martini) 框架，但又不想框架中有太多魔幻性的特征，那 Negroni 就是你的菜了，相信它非常适合你。


语言翻译:
* [Português Brasileiro (pt_BR)](translations/README_pt_br.md)
* [简体中文 (zh_CN)](translations/README_zh_cn.md)

## 入门指导

当安装了 Go 语言并设置好了 [GOPATH](http://golang.org/doc/code.html#GOPATH) 后，新建你第一个`.go` 文件，我们叫它 `server.go` 吧。

~~~ go
package main

import (
  "github.com/urfave/negroni"
  "net/http"
  "fmt"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic()
  n.UseHandler(mux)
  n.Run(":3000")
}
~~~

然后安装 Negroni 包（它依赖 **Go 1.1** 或更高的版本）：
~~~
go get github.com/urfave/negroni
~~~

然后运行刚建好的 server.go 文件:
~~~
go run server.go
~~~

这时一个 Go `net/http` Web 服务器就跑在 `localhost:3000` 上，使用浏览器打开 `localhost:3000` 可以看到输出结果。

## 需要你的贡献
如果你有问题或新想法，请到[邮件群组](https://groups.google.com/forum/#!forum/negroni-users)里反馈，GitHub issues 是专门给提交 bug 报告和 pull 请求用途的，欢迎你的参与。

## Negroni 是一个框架吗？
Negroni **不**是一个框架，它是为了方便使用 `net/http` 而设计的一个库而已。

## 路由呢？
Negroni 没有带路由功能，使用 Negroni 时，需要找一个适合你的路由。不过好在 Go 社区里已经有相当多可用的路由，Negroni 更喜欢和那些完全支持 `net/http` 库的路由组合使用，比如，结合 [Gorilla Mux](http://github.com/gorilla/mux) 使用像这样：

~~~ go
router := mux.NewRouter()
router.HandleFunc("/", HomeHandler)

n := negroni.New(Middleware1, Middleware2)
// Or use a middleware with the Use() function
n.Use(Middleware3)
// router goes last
n.UseHandler(router)

n.Run(":3000")
~~~

## `negroni.Classic()` 经典实例
`negroni.Classic()` 提供一些默认的中间件，这些中间件在多数应用都很有用。

* `negroni.Recovery` - 异常（恐慌）恢复中间件
* `negroni.Logging` - 请求 / 响应 log 日志中间件
* `negroni.Static` - 静态文件处理中间件，默认目录在 "public" 下.

`negroni.Classic()` 让你一开始就非常容易上手 Negroni ，并使用它那些通用的功能。

## Handlers (处理器)
Negroni 提供双向的中间件机制，这个特征很棒，都是得益于 `negroni.Handler` 这个接口。

~~~ go
type Handler interface {
  ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
~~~

如果一个中间件没有写入 ResponseWriter 响应，它会在中间件链里调用下一个 `http.HandlerFunc` 执行下去， 它可以这么优雅的使用。如下：

~~~ go
func MyMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  // do some stuff before
  next(rw, r)
  // do some stuff after
}
~~~

你也可以用 `Use` 函数把这些 `http.Handler` 处理器引进到处理器链上来：

~~~ go
n := negroni.New()
n.Use(negroni.HandlerFunc(MyMiddleware))
~~~

你还可以使用 `http.Handler`(s) 把 `http.Handler` 处理器引进来。

~~~ go
n := negroni.New()

mux := http.NewServeMux()
// map your routes

n.UseHandler(mux)

n.Run(":3000")
~~~

## `Run()`
Negroni 提供一个很好用的函数叫 `Run` ，把地址字符串传人该函数，即可实现很地道的 [http.ListenAndServe](http://golang.org/pkg/net/http#ListenAndServe) 函数功能了。

~~~ go
n := negroni.Classic()
// ...
log.Fatal(http.ListenAndServe(":8080", n))
~~~

## 特定路由中间件
如果你需要群组路由功能，需要借助特定的路由中间件完成，做法很简单，只需建立一个新 Negroni 实例，传人路由处理器里即可。

~~~ go
router := mux.NewRouter()
adminRoutes := mux.NewRouter()
// add admin routes here

// Create a new negroni for the admin middleware
router.Handle("/admin", negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(adminRoutes),
))
~~~

## 第三方中间件

以下的兼容 Negroni 的中间件列表，如果你也有兼容 Negroni 的中间件，可以提交到这个列表来交换链接，我们很乐意做这样有益的事情。


|    中间件    |    作者    |    描述     |
| -------------|------------|-------------|
| [RestGate](https://github.com/pjebs/restgate) | [Prasanga Siripala](https://github.com/pjebs) | REST API 接口的安全认证 |
| [Graceful](https://github.com/stretchr/graceful) | [Tyler Bunnell](https://github.com/tylerb) | 优雅关闭 HTTP 的中间件 |
| [secure](https://github.com/unrolled/secure) | [Cory Jacobsen](https://github.com/unrolled) | Middleware that implements a few quick security wins |
| [JWT Middleware](https://github.com/auth0/go-jwt-middleware) | [Auth0](https://github.com/auth0) | Middleware checks for a JWT on the `Authorization` header on incoming requests and decodes it|
| [binding](https://github.com/mholt/binding) | [Matt Holt](https://github.com/mholt) | HTTP 请求数据注入到 structs 实体|
| [logrus](https://github.com/meatballhat/negroni-logrus) | [Dan Buch](https://github.com/meatballhat) | 基于 Logrus-based logger 日志 |
| [render](https://github.com/unrolled/render) | [Cory Jacobsen](https://github.com/unrolled) | 渲染 JSON, XML and HTML 中间件 |
| [gorelic](https://github.com/jingweno/negroni-gorelic) | [Jingwen Owen Ou](https://github.com/jingweno) | New Relic agent for Go runtime |
| [gzip](https://github.com/phyber/negroni-gzip) | [phyber](https://github.com/phyber) | 响应流 GZIP 压缩 |
| [oauth2](https://github.com/goincremental/negroni-oauth2) | [David Bochenski](https://github.com/bochenski) | oAuth2 中间件 |
| [sessions](https://github.com/goincremental/negroni-sessions) | [David Bochenski](https://github.com/bochenski) | Session 会话管理 |
| [permissions2](https://github.com/xyproto/permissions2) | [Alexander Rødseth](https://github.com/xyproto) | Cookies， 用户和权限 |
| [onthefly](https://github.com/xyproto/onthefly) | [Alexander Rødseth](https://github.com/xyproto) | 快速生成 TinySVG， HTML and CSS 中间件 |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | [Cross Origin Resource Sharing](http://www.w3.org/TR/cors/) (CORS) support |
| [xrequestid](https://github.com/pilu/xrequestid) | [Andrea Franz](https://github.com/pilu) | 给每个请求指定一个随机 X-Request-Id 头的中间件 |
| [VanGoH](https://github.com/auroratechnologies/vangoh) | [Taylor Wrobel](https://github.com/twrobel3) | Configurable [AWS-Style](http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html) 基于 HMAC 鉴权认证的中间件 |
| [stats](https://github.com/thoas/stats) | [Florent Messa](https://github.com/thoas) | 检测 web 应用当前运行状态信息 （响应时间等等。） |

## 范例
[Alexander Rødseth](https://github.com/xyproto) 创建的 [mooseware](https://github.com/xyproto/mooseware) 是一个写兼容 Negroni 中间件的处理器骨架的范例。

## 即时编译
[gin](https://github.com/codegangsta/gin) 和 [fresh](https://github.com/pilu/fresh) 这两个应用是即时编译的 Negroni 工具，推荐用户开发的时候使用。

## Go & Negroni 初学者必读推荐

* [在中间件中使用上下文把消息传递给后端处理器](http://elithrar.github.io/article/map-string-interface/)
* [了解中间件](http://mattstauffer.co/blog/laravel-5.0-middleware-replacing-filters)

## 关于

Negroni 由 [Code Gangsta](http://codegangsta.io/) 主导设计开发完成 