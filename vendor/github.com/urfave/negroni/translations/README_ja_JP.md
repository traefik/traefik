# Negroni [![GoDoc](https://godoc.org/github.com/urfave/negroni?status.svg)](http://godoc.org/github.com/urfave/negroni) [![wercker status](https://app.wercker.com/status/13688a4a94b82d84a0b8d038c4965b61/s "wercker status")](https://app.wercker.com/project/bykey/13688a4a94b82d84a0b8d038c4965b61) [![codebeat](https://codebeat.co/badges/47d320b1-209e-45e8-bd99-9094bc5111e2)](https://codebeat.co/projects/github-com-urfave-negroni)

NegroniはGoによるWeb ミドルウェアへの慣用的なアプローチです。
軽量で押し付けがましい作法は無く、また`net/http`ハンドラの使用を推奨しています。

[Martini](https://github.com/go-martini/martini) の思想は気に入っているが、多くの魔法を含みすぎていると感じている方に、このNegroni はよく馴染むでしょう。

## はじめに

Goをインストールし、[GOPATH](http://golang.org/doc/code.html#GOPATH)の設定を行った後、`.go`ファイルを作りましょう。これを`server.go`とします。

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

  n := negroni.Classic() // Includes some default middlewares
  n.UseHandler(mux)

  http.ListenAndServe(":3000", n)
}
```

Negroni パッケージをインストールします (**NOTE**: &gt;= **go 1.1** 以上のバージョンが必要です):

```
go get github.com/urfave/negroni
```

インストールが完了したら、サーバーを起動しましょう。

```
go run server.go
```

すると、Go標準パッケージの `net/http` によるWebサーバーが`localhost:3000` で起動します。

## Negroni はWeb Application Framework ですか？

Negroni はrevel やmartini のような**フレームワークではありません**。 Negroniは `net/http`と直接結びついて動作する、ミドルウェアにフォーカスされたライブラリです。


## ルーティングの機能はありますか？

Negroni にルーティングの機能はありません。Goコミュニティには既に幾つかの優れたルーティングのライブラリが存在しており、Negroni は`net/http`と互換性のあるライブラリと協調して動作するように設計されています。例えば、[Gorilla Mux]と連携すると以下のようになります。

``` go
router := mux.NewRouter()
router.HandleFunc("/", HomeHandler)

n := negroni.New(Middleware1, Middleware2)
// Or use a middleware with the Use() function
n.Use(Middleware3)
// router goes last
n.UseHandler(router)

http.ListenAndServe(":3001", n)
```

## `negroni.Classic()`

`negroni.Classic()` は、多くのアプリケーションで役に立つミドルウェアを提供します

* [`negroni.Recovery`](#recovery) - Panic Recovery Middleware.
* [`negroni.Logger`](#logger) - Request/Response Logger Middleware.
* [`negroni.Static`](#static) - "public"ディレクトリの静的ファイルの処理

これらはNegroni の便利な機能を利用し始めるのをとても簡単にしてくれます。

## ハンドラ

Negroniは双方向のミドルウェアのフローを提供します。これは、`negroni.Handler` インターフェースを通じて行われます。

``` go
type Handler interface {
  ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
```

ミドルウェアが既に`ResponseWriter`に書き込み処理を行っていない場合、次のミドルウェア・ハンドラを動かすために、チェーン内の次の`http.HandlerFunc`を呼び出す必要があります。


``` go
func MyMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  // 前処理
  next(rw, r)
  // 後処理
}
```

この時、`MyMiddleware`を`Use` 関数によってハンドラチェーンに割り当てることができます。

``` go
n := negroni.New()
n.Use(negroni.HandlerFunc(MyMiddleware))
```

また、標準パッケージに備わっている`http.Handler`をハンドラチェーンに割り当てることもできます。

``` go
n := negroni.New()

mux := http.NewServeMux()
// ルーティングの処理

n.UseHandler(mux)

http.ListenAndServe(":3000", n)
```

## `Run()`

`Run` はアドレスの文字列を受け取り、[`http.ListenAndServe`](https://godoc.org/net/http#ListenAndServe)と同様にサーバーを起動します。

<!-- { "interrupt": true } -->
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

通常、`net/http` を使用し、ハンドラとして`Negroni`を渡すことになります。
以下により柔軟性のあるサンプルを示します。

<!-- { "interrupt": true } -->
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

  n := negroni.Classic()
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

## Route Specific Middleware

あるルーティンググループにおいて、実行する必要のあるミドルウェアがある場合、
新しいNegroni のインスタンスを作成し、ルーティングハンドラとして使用することができます。

``` go
router := mux.NewRouter()
adminRoutes := mux.NewRouter()
// admin 関連のルーティングをココに記述

// admin のミドルウェアとして、Negroni インスタンスを作成
router.PathPrefix("/admin").Handler(negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(adminRoutes),
))
```

もし[Gorilla Mux]を利用する場合、サブルーターを使うサンプルは以下の通りです。

``` go
router := mux.NewRouter()
subRouter := mux.NewRouter().PathPrefix("/subpath").Subrouter().StrictSlash(true)
subRouter.HandleFunc("/", someSubpathHandler) // "/subpath/"
subRouter.HandleFunc("/:id", someSubpathHandler) // "/subpath/:id"

// "/subpath" is necessary to ensure the subRouter and main router linkup
router.PathPrefix("/subpath").Handler(negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(subRouter),
))
```

## Bundled Middleware

### Static

このミドルウェアは、ファイルシステム上のファイルをサーバーからクライアントに送信します。もし指定されたファイルが存在しない場合、次のミドルウェアにリクエストの処理を依頼します。存在しないファイルへのアクセスに対して`404 Not Found`を返したい場合、 [http.FileServer](https://golang.org/pkg/net/http/#FileServer) をハンドラとして利用すべきです。

Example:

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

  // Example of using a http.FileServer if you want "server-like" rather than "middleware" behavior
  // mux.Handle("/public", http.FileServer(http.Dir("/home/public")))

  n := negroni.New()
  n.Use(negroni.NewStatic(http.Dir("/tmp")))
  n.UseHandler(mux)

  http.ListenAndServe(":3002", n)
}
```

まず、`/tmp` ディレクトリからファイルをクライアントに送ろうとしますが、指定されたファイルがファイルシステム上に存在しない場合、プロキシは次のハンドラを呼び出します。

### Recovery

このミドルウェアは、`panic`を受け取ると、`500 internal Server Error` をレスポンスします。
もし他のミドルウェアが既に応答処理を行い、クライアントにHTTP レスポンスが帰っている場合、このミドルウェアは失敗します。

Example:

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

上記のコードは、 `500 Internal Server Error` を各リクエストごとに返します。
また、スタックトレースをログに出力するだけでなく、`PrintStack`が`true`に設定されている場合、クライアントにスタックトレースを出力します。（デフォルトで`true`に設定されています。）

## Logger

このミドルウェアは、送られてきた全てのリクエストとレスポンスを記録します。

Example:

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

各リクエストごとに、以下のようなログが出力されます。

```
[negroni] Started GET /
[negroni] Completed 200 OK in 145.446µs
```


## サードパーティ製のミドルウェア

Negroni と互換性のあるミドルウェアの一覧です。あなたが作ったミドルウェアをここに追加してもらっても構いません（PRを送って下さい）

**注意**: ここの一覧は古くなっている可能性があります。英語版のREADME.md を適宜参照して下さい。


| ミドルウェア名 | 作者 | 概要 |
| -----------|--------|-------------|
| [binding](https://github.com/mholt/binding) | [Matt Holt](https://github.com/mholt) | Data binding from HTTP requests into structs |
| [cloudwatch](https://github.com/cvillecsteele/negroni-cloudwatch) | [Colin Steele](https://github.com/cvillecsteele) | AWS cloudwatch metrics middleware |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | [Cross Origin Resource Sharing](http://www.w3.org/TR/cors/) (CORS) support |
| [delay](https://github.com/jeffbmartinez/delay) | [Jeff Martinez](https://github.com/jeffbmartinez) | Add delays/latency to endpoints. Useful when testing effects of high latency |
| [gorelic](https://github.com/jingweno/negroni-gorelic) | [Jingwen Owen Ou](https://github.com/jingweno) | New Relic agent for Go runtime |
| [Graceful](https://github.com/tylerb/graceful) | [Tyler Bunnell](https://github.com/tylerb) | Graceful HTTP Shutdown |
| [gzip](https://github.com/phyber/negroni-gzip) | [phyber](https://github.com/phyber) | GZIP response compression |
| [JWT Middleware](https://github.com/auth0/go-jwt-middleware) | [Auth0](https://github.com/auth0) | Middleware checks for a JWT on the `Authorization` header on incoming requests and decodes it|
| [logrus](https://github.com/meatballhat/negroni-logrus) | [Dan Buch](https://github.com/meatballhat) | Logrus-based logger |
| [oauth2](https://github.com/goincremental/negroni-oauth2) | [David Bochenski](https://github.com/bochenski) | oAuth2 middleware |
| [onthefly](https://github.com/xyproto/onthefly) | [Alexander Rødseth](https://github.com/xyproto) | Generate TinySVG, HTML and CSS on the fly |
| [permissions2](https://github.com/xyproto/permissions2) | [Alexander Rødseth](https://github.com/xyproto) | Cookies, users and permissions |
| [prometheus](https://github.com/zbindenren/negroni-prometheus) | [Rene Zbinden](https://github.com/zbindenren) | Easily create metrics endpoint for the [prometheus](http://prometheus.io) instrumentation tool |
| [render](https://github.com/unrolled/render) | [Cory Jacobsen](https://github.com/unrolled) | Render JSON, XML and HTML templates |
| [RestGate](https://github.com/pjebs/restgate) | [Prasanga Siripala](https://github.com/pjebs) | Secure authentication for REST API endpoints |
| [secure](https://github.com/unrolled/secure) | [Cory Jacobsen](https://github.com/unrolled) | Middleware that implements a few quick security wins |
| [sessions](https://github.com/goincremental/negroni-sessions) | [David Bochenski](https://github.com/bochenski) | Session Management |
| [stats](https://github.com/thoas/stats) | [Florent Messa](https://github.com/thoas) | Store information about your web application (response time, etc.) |
| [VanGoH](https://github.com/auroratechnologies/vangoh) | [Taylor Wrobel](https://github.com/twrobel3) | Configurable [AWS-Style](http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html) HMAC authentication middleware |
| [xrequestid](https://github.com/pilu/xrequestid) | [Andrea Franz](https://github.com/pilu) | Middleware that assigns a random X-Request-Id header to each request |

## Examples

[Alexander Rødseth](https://github.com/xyproto) created
[mooseware](https://github.com/xyproto/mooseware), a skeleton for writing a
Negroni middleware handler.

## Live code reload?

[gin](https://github.com/codegangsta/gin) and
[fresh](https://github.com/pilu/fresh) both live reload negroni apps.

## Go や Negroni の初心者にオススメの参考資料（英語）

* [Using a Context to pass information from middleware to end handler](http://elithrar.github.io/article/map-string-interface/)
* [Understanding middleware](https://mattstauffer.co/blog/laravel-5.0-middleware-filter-style)

## Negroni について

Negroni は他ならぬ[Code
Gangsta](https://codegangsta.io/)によって異常なまでにデザインされた素晴らしいライブラリです。

[Gorilla Mux]: https://github.com/gorilla/mux
[`http.FileSystem`]: https://godoc.org/net/http#FileSystem
