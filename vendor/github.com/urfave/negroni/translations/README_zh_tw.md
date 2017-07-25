# Negroni(尼格龍尼)
[![GoDoc](https://godoc.org/github.com/urfave/negroni?status.svg)](http://godoc.org/github.com/urfave/negroni)
[![Build Status](https://travis-ci.org/urfave/negroni.svg?branch=master)](https://travis-ci.org/urfave/negroni)
[![codebeat](https://codebeat.co/badges/47d320b1-209e-45e8-bd99-9094bc5111e2)](https://codebeat.co/projects/github-com-urfave-negroni)
[![codecov](https://codecov.io/gh/urfave/negroni/branch/master/graph/badge.svg)](https://codecov.io/gh/urfave/negroni)

**注意:** 本函式庫原來自於
`github.com/codegangsta/negroni` -- Github會自動將連線轉到本連結, 但我們建議你更新一下參照.

尼格龍尼符合Go的web 中介器特性. 精簡、非侵入式、鼓勵使用 `net/http`  Handler.

如果你喜歡[Martini](http://github.com/go-martini/martini), 但覺得這其中包太多神奇的功能, 那麼尼格龍尼會是你的最佳選擇.

其他語言:
* [German (de_DE)](translations/README_de_de.md)
* [Português Brasileiro (pt_BR)](translations/README_pt_br.md)
* [简体中文 (zh_cn)](translations/README_zh_cn.md)
* [繁體中文 (zh_tw)](translations/README_zh_tw.md)
* [日本語 (ja_JP)](translations/README_ja_JP.md)

## 入門

安裝完Go且設好[GOPATH](http://golang.org/doc/code.html#GOPATH), 建立你的第一個`.go`檔. 可以命名為`server.go`.

``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic() // 導入一些預設中介器
  n.UseHandler(mux)

  http.ListenAndServe(":3000", n)
}
```

安裝尼格龍尼套件 (最低需求為**go 1.1**或更高版本):
```
go get github.com/urfave/negroni
```

執行伺服器:
```
go run server.go
```

你現在起了一個Go的net/http網頁伺服器在`localhost:3000`.

## 有問題?
如果你有問題或新功能建議, [到這郵件群組討論](https://groups.google.com/forum/#!forum/negroni-users). 尼格龍尼在GitHub上的issues專欄是專門用來回報bug跟pull requests.

## 尼格龍尼是個framework嗎?
尼格龍尼**不是**framework, 是個設計用來直接使用net/http的library.

## 路由?
尼格龍尼是BYOR (Bring your own Router, 帶給你自訂路由). 在Go社群已經有大量可用的http路由器, 尼格龍尼試著做好完全支援`net/http`, 例如與[Gorilla Mux](http://github.com/gorilla/mux)整合:

``` go
router := mux.NewRouter()
router.HandleFunc("/", HomeHandler)

n := negroni.New(Middleware1, Middleware2)
// 或在Use()函式中使用中介器
n.Use(Middleware3)
// 路由器放最後
n.UseHandler(router)

http.ListenAndServe(":3001", n)
```

## `negroni.Classic()`
`negroni.Classic()` 提供一些好用的預設中介器:

* [`negroni.Recovery`](https://github.com/urfave/negroni#recovery) - Panic 還原中介器
* [`negroni.Logging`](https://github.com/urfave/negroni#logger) - Request/Response 紀錄中介器
* [`negroni.Static`](https://github.com/urfave/negroni#static) - 在"public"目錄下的靜態檔案服務

尼格龍尼的這些功能讓你開發變得很簡單.

## 處理器(Handlers)
尼格龍尼提供一個雙向中介器的機制, 介面為`negroni.Handler`:

``` go
type Handler interface {
  ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
```

如果中介器沒有寫入ResponseWriter, 會呼叫通道裡面的下個`http.HandlerFunc`讓給中介處理器. 可以被用來做良好的應用:

``` go
func MyMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  // 在這之前做一些事
  next(rw, r)
  // 在這之後做一些事
}
```

然後你可以透過 `Use` 函數對應到處理器的通道:

``` go
n := negroni.New()
n.Use(negroni.HandlerFunc(MyMiddleware))
```

你也可以對應原始的 `http.Handler`:

``` go
n := negroni.New()

mux := http.NewServeMux()
// map your routes

n.UseHandler(mux)

http.ListenAndServe(":3000", n)
```

## `Run()`
尼格龍尼有一個很好用的函數`Run`, `Run`接收addr字串辨識[http.ListenAndServe](http://golang.org/pkg/net/http#ListenAndServe).

``` go
package main

import (
  "github.com/urfave/negroni"
)

func main() {
  n := negroni.Classic()
  n.Run(":8080")
}
```

In general, you will want to use net/http methods and pass negroni as a Handler, as this is more flexible, e.g.:
一般來說, 你會希望使用 `net/http` 方法, 並且將尼格龍尼當作處理器傳入, 這相對起來彈性比較大, 例如：

``` go
package main

import (
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic() // 導入一些預設中介器
  n.UseHandler(mux)

  s := &http.Server{
    Addr:           ":8080",
    Handler:        n,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }
  log.Fatal(s.ListenAndServe())
}
```

## 路由特有中介器
如果你有一群路由需要執行特別的中介器, 你可以簡單的建立一個新的尼格龍尼實體當作路由處理器.

``` go
router := mux.NewRouter()
adminRoutes := mux.NewRouter()
// 在這裡新增管理用的路由

// 為管理中介器建立一個新的尼格龍尼
router.Handle("/admin", negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(adminRoutes),
))
```

如果你使用 [Gorilla Mux](https://github.com/gorilla/mux), 下方是一個使用 subrounter 的例子：

``` go
router := mux.NewRouter()
subRouter := mux.NewRouter().PathPrefix("/subpath").Subrouter().StrictSlash(true)
subRouter.HandleFunc("/", someSubpathHandler) // "/subpath/"
subRouter.HandleFunc("/:id", someSubpathHandler) // "/subpath/:id"

// "/subpath" 是用來保證subRouter與主要路由連結的必要參數
router.PathPrefix("/subpath").Handler(negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(subRouter),
))
```

`With()` 可被用來降低在跨路由分享時多餘的中介器.

``` go
router := mux.NewRouter()
apiRoutes := mux.NewRouter()
// 在此新增API路由
webRoutes := mux.NewRouter()
// 在此新增Web路由

// 建立通用中介器來跨路由分享
common := negroni.New(
  Middleware1,
  Middleware2,
)

// 為API中介器建立新的negroni
// 使用通用中介器作底
router.PathPrefix("/api").Handler(common.With(
  APIMiddleware1,
  negroni.Wrap(apiRoutes),
))
// 為Web中介器建立新的negroni
// 使用通用中介器作底
router.PathPrefix("/web").Handler(common.With(
  WebMiddleware1,
  negroni.Wrap(webRoutes),
))
```


## 內建中介器

### 靜態

本中介器會在檔案系統上服務檔案. 若檔案不存在, 會將流量導(proxy)到下個中介器.
如果你想要返回`404 File Not Found`給檔案不存在的請求, 請使用[http.FileServer](https://golang.org/pkg/net/http/#FileServer)
作為處理器.

範例:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  // http.FileServer的使用範例, 若你預期要"像伺服器"而非"中介器"的行為
  // mux.Handle("/public", http.FileServer(http.Dir("/home/public")))

  n := negroni.New()
  n.Use(negroni.NewStatic(http.Dir("/tmp")))
  n.UseHandler(mux)

  http.ListenAndServe(":3002", n)
}
```

從`/tmp`目錄開始服務檔案 但如果請求的檔案在檔案系統中不符合, 代理會
呼叫下個處理器.

### 恢復

本中介器接收`panic`跟錯誤代碼`500`的回應. 如果其他任何中介器寫了回應
的HTTP代碼或內容的話, 中介器會無法順利地傳送500給用戶端, 因為用戶端
已經收到HTTP的回應代碼. 另外, 可以掛載`ErrorHandlerFunc`來回報500
的錯誤到錯誤回報系統, 如: Sentry或Airbrake.

範例:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    panic("oh no")
  })

  n := negroni.New()
  n.Use(negroni.NewRecovery())
  n.UseHandler(mux)

  http.ListenAndServe(":3003", n)
}
```


將回傳`500 Internal Server Error`到每個結果. 也會把結果紀錄到堆疊追蹤,
`PrintStack`設成`true`(預設值)的話也會印到註冊者.

加錯誤處理器的範例:

``` go
package main

import (
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    panic("oh no")
  })

  n := negroni.New()
  recovery := negroni.NewRecovery()
  recovery.ErrorHandlerFunc = reportToSentry
  n.Use(recovery)
  n.UseHandler(mux)

  http.ListenAndServe(":3003", n)
}

func reportToSentry(error interface{}) {
    // 在這寫些程式回報錯誤給Sentry
}
```


## Logger

本中介器紀錄各個進入的請求與回應.

範例:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.New()
  n.Use(negroni.NewLogger())
  n.UseHandler(mux)

  http.ListenAndServe(":3004", n)
}
```

在每個請求印的紀錄會看起來像:

```
[negroni] Started GET /
[negroni] Completed 200 OK in 145.446µs
```

## 第三方中介器

以下清單是目前可用於尼格龍尼的中介器. 如果你自己手癢做了一個, 請別吝嗇自己把連結貼在下面吧:

| 中介器 | 作者 | 說明 |
| -----------|--------|-------------|
| [binding](https://github.com/mholt/binding) | [Matt Holt](https://github.com/mholt) | 把HTTP請求的資料榜定到structs |
| [cloudwatch](https://github.com/cvillecsteele/negroni-cloudwatch) | [Colin Steele](https://github.com/cvillecsteele) | AWS CloudWatch 矩陣的中介器 |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | 支援[Cross Origin Resource Sharing](http://www.w3.org/TR/cors/)(CORS) |
| [csp](https://github.com/awakenetworks/csp) | [Awake Networks](https://github.com/awakenetworks) | 支援[Content Security Policy](https://www.w3.org/TR/CSP2/)(CSP) |
| [delay](https://github.com/jeffbmartinez/delay) | [Jeff Martinez](https://github.com/jeffbmartinez) | 為endpoints增加延遲時間. 在測試嚴重網路延遲的效應時好用 |
| [New Relic Go Agent](https://github.com/yadvendar/negroni-newrelic-go-agent) | [Yadvendar Champawat](https://github.com/yadvendar) | 官網 [New Relic Go Agent](https://github.com/newrelic/go-agent) (目前正在測試階段)  |
| [gorelic](https://github.com/jingweno/negroni-gorelic) | [Jingwen Owen Ou](https://github.com/jingweno) | New Relic agent for Go runtime |
| [Graceful](https://github.com/tylerb/graceful) | [Tyler Bunnell](https://github.com/tylerb) | 優雅地關閉HTTP |
| [gzip](https://github.com/phyber/negroni-gzip) | [phyber](https://github.com/phyber) | GZIP資源壓縮 |
| [JWT Middleware](https://github.com/auth0/go-jwt-middleware) | [Auth0](https://github.com/auth0) | Middleware 檢查JWT在`Authorization` header on incoming requests and decodes it|
| [logrus](https://github.com/meatballhat/negroni-logrus) | [Dan Buch](https://github.com/meatballhat) | 基於Logrus的紀錄器 |
| [oauth2](https://github.com/goincremental/negroni-oauth2) | [David Bochenski](https://github.com/bochenski) | oAuth2中介器 |
| [onthefly](https://github.com/xyproto/onthefly) | [Alexander Rødseth](https://github.com/xyproto) | 一秒產生TinySVG, HTML, CSS |
| [permissions2](https://github.com/xyproto/permissions2) | [Alexander Rødseth](https://github.com/xyproto) | Cookies與使用者權限配套 |
| [prometheus](https://github.com/zbindenren/negroni-prometheus) | [Rene Zbinden](https://github.com/zbindenren) | 簡易建立矩陣端點給[prometheus](http://prometheus.io)建構工具 |
| [render](https://github.com/unrolled/render) | [Cory Jacobsen](https://github.com/unrolled) |  JSON, XML, HTML樣板的渲染 |
| [RestGate](https://github.com/pjebs/restgate) | [Prasanga Siripala](https://github.com/pjebs) | REST API端點安全認證 |
| [secure](https://github.com/unrolled/secure) | [Cory Jacobsen](https://github.com/unrolled) | 簡易安全中介器 |
| [sessions](https://github.com/goincremental/negroni-sessions) | [David Bochenski](https://github.com/bochenski) | Session 管理 |
| [stats](https://github.com/thoas/stats) | [Florent Messa](https://github.com/thoas) | 儲存關於網頁應用的資訊(回應時間之類) |
| [VanGoH](https://github.com/auroratechnologies/vangoh) | [Taylor Wrobel](https://github.com/twrobel3) | 可設定的[AWS風格](http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html) HMAC認證中介器 |
| [xrequestid](https://github.com/pilu/xrequestid) | [Andrea Franz](https://github.com/pilu) | 在每個request指定一個隨機X-Request-Id header的中介器 |
| [mgo session](https://github.com/joeljames/nigroni-mgo-session) | [Joel James](https://github.com/joeljames) | 處理在每個請求建立與關閉mgo sessions |
| [digits](https://github.com/bamarni/digits) | [Bilal Amarni](https://github.com/bamarni) | 處理[Twitter Digits](https://get.digits.com/)的認證 |

## 應用範例

[Alexander Rødseth](https://github.com/xyproto)所建
[mooseware](https://github.com/xyproto/mooseware)用來寫尼格龍尼中介處理器的骨架

## Live code reload?

[gin](https://github.com/codegangsta/gin)和
[fresh](https://github.com/pilu/fresh)兩個尼格龍尼即時重載的應用.

## Go & 尼格龍尼初學者必讀

* [使用Context將資訊從中介器送到處理器](http://elithrar.github.io/article/map-string-interface/)
* [理解中介器](https://mattstauffer.co/blog/laravel-5.0-middleware-filter-style)

## 關於

尼格龍尼正是[Code Gangsta](https://codegangsta.io/)的執著設計.

[Gorilla Mux]: https://github.com/gorilla/mux
[`http.FileSystem`]: https://godoc.org/net/http#FileSystem

譯者: Festum Qin (Festum@G.PL)
