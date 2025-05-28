import { Middleware } from 'hooks/use-resource-detail'

export const parseMiddlewareType = (middleware: Middleware): string | undefined => {
  if (middleware.plugin) {
    const pluginObject = middleware.plugin || {}
    const [pluginName] = Object.keys(pluginObject)

    if (pluginName) {
      return pluginName
    }
  }

  return middleware.type
}
