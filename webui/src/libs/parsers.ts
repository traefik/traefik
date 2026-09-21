export const parseMiddlewareType = (middleware: Middleware.Details): string | undefined => {
  if (middleware.plugin) {
    const pluginObject = middleware.plugin || {}
    const [pluginName] = Object.keys(pluginObject)

    if (pluginName) {
      return pluginName
    }
  }

  return middleware.type
}
