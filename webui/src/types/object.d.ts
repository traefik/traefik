declare namespace Object {
  type JSONObject = {
    [x: string]: string | number
  }

  type ValuesMapType = {
    [key: string]: string | number | JSONObject
  }
}
