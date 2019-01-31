/*Package correlation is a HTTP middleware that adds correlation ids to incoming requests.

  package main

  import (
      "net/http"

      "gitlab.com/JanMa/correlation"
  )

  var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("hello world"))
  })

  func main() {
      correlationMiddleware := correlation.New(correlation.Options{
          CorrelationIDType: correlation.UUID,
      })

      http.ListenAndServe(":8080", correlationMiddleware.Handler(myHandler))
  }
*/
package correlation
