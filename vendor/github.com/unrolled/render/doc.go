/*Package render is a package that provides functionality for easily rendering JSON, XML, binary data, and HTML templates.

  package main

  import (
      "encoding/xml"
      "net/http"

      "github.com/unrolled/render"  // or "gopkg.in/unrolled/render.v1"
  )

  type ExampleXml struct {
      XMLName xml.Name `xml:"example"`
      One     string   `xml:"one,attr"`
      Two     string   `xml:"two,attr"`
  }

  func main() {
      r := render.New()
      mux := http.NewServeMux()

      mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
          w.Write([]byte("Welcome, visit sub pages now."))
      })

      mux.HandleFunc("/data", func(w http.ResponseWriter, req *http.Request) {
          r.Data(w, http.StatusOK, []byte("Some binary data here."))
      })

      mux.HandleFunc("/text", func(w http.ResponseWriter, req *http.Request) {
          r.Text(w, http.StatusOK, "Plain text here")
      })

      mux.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
          r.JSON(w, http.StatusOK, map[string]string{"hello": "json"})
      })

      mux.HandleFunc("/jsonp", func(w http.ResponseWriter, req *http.Request) {
          r.JSONP(w, http.StatusOK, "callbackName", map[string]string{"hello": "jsonp"})
      })

      mux.HandleFunc("/xml", func(w http.ResponseWriter, req *http.Request) {
          r.XML(w, http.StatusOK, ExampleXml{One: "hello", Two: "xml"})
      })

      mux.HandleFunc("/html", func(w http.ResponseWriter, req *http.Request) {
          // Assumes you have a template in ./templates called "example.tmpl".
          // $ mkdir -p templates && echo "<h1>Hello HTML world.</h1>" > templates/example.tmpl
          r.HTML(w, http.StatusOK, "example", nil)
      })

      http.ListenAndServe("0.0.0.0:3000", mux)
  }
*/
package render
