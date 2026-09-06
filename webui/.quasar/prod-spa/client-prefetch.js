/* eslint-disable */
/**
 * THIS FILE IS GENERATED AUTOMATICALLY.
 * DO NOT EDIT.
 *
 * You are probably looking on adding startup/initialization code.
 * Use "quasar new boot <name>" and add it there.
 * One boot file per concern. Then reference the file(s) in quasar.config file > boot:
 * boot: ['file', ...] // do not add ".js" extension to it.
 *
 * Boot files are your "main.js"
 **/

import { LoadingBar } from 'quasar'



import App from 'app/src/App.vue'
let appPrefetch = typeof App.preFetch === 'function'
  ? App.preFetch
  : (
    // Class components return the component options (and the preFetch hook) inside __c property
    App.__c !== void 0 && typeof App.__c.preFetch === 'function'
      ? App.__c.preFetch
      : false
    )


function getMatchedComponents (to, router) {
  const route = to
    ? (to.matched ? to : router.resolve(to).route)
    : router.currentRoute.value

  if (!route) { return [] }

  const matched = route.matched.filter(m => m.components !== void 0)

  if (matched.length === 0) { return [] }

  return Array.prototype.concat.apply([], matched.map(m => {
    return Object.keys(m.components).map(key => {
      const comp = m.components[key]
      return {
        path: m.path,
        c: comp
      }
    })
  }))
}

export function addPreFetchHooks ({ router, store, publicPath }) {
  // Add router hook for handling preFetch.
  // Doing it after initial route is resolved so that we don't double-fetch
  // the data that we already have. Using router.beforeResolve() so that all
  // async components are resolved.
  router.beforeResolve((to, from, next) => {
    const
      urlPath = window.location.href.replace(window.location.origin, ''),
      matched = getMatchedComponents(to, router),
      prevMatched = getMatchedComponents(from, router)

    let diffed = false
    const preFetchList = matched
      .filter((m, i) => {
        return diffed || (diffed = (
          !prevMatched[i] ||
          prevMatched[i].c !== m.c ||
          m.path.indexOf('/:') > -1 // does it has params?
        ))
      })
      .filter(m => m.c !== void 0 && (
        typeof m.c.preFetch === 'function'
        // Class components return the component options (and the preFetch hook) inside __c property
        || (m.c.__c !== void 0 && typeof m.c.__c.preFetch === 'function')
      ))
      .map(m => m.c.__c !== void 0 ? m.c.__c.preFetch : m.c.preFetch)

    
    if (appPrefetch !== false) {
      preFetchList.unshift(appPrefetch)
      appPrefetch = false
    }
    

    if (preFetchList.length === 0) {
      return next()
    }

    let hasRedirected = false
    const redirect = url => {
      hasRedirected = true
      next(url)
    }
    const proceed = () => {
      
      LoadingBar.stop()
      
      if (hasRedirected === false) { next() }
    }

    
    LoadingBar.start()
    

    preFetchList.reduce(
      (promise, preFetch) => promise.then(() => hasRedirected === false && preFetch({
        store,
        currentRoute: to,
        previousRoute: from,
        redirect,
        urlPath,
        publicPath
      })),
      Promise.resolve()
    )
    .then(proceed)
    .catch(e => {
      console.error(e)
      proceed()
    })
  })
}
