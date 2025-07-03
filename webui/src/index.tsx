import { createRoot } from 'react-dom/client'

import App from './App'

async function enableMocking() {
  if (import.meta.env.MODE !== 'development') {
    return
  }

  const { worker } = await import('./mocks/browser')

  // `worker.start()` returns a Promise that resolves
  // once the Service Worker is up and ready to intercept requests.
  return worker.start()
}

enableMocking().then(() => {
  const container = document.getElementById('root')
  const root = createRoot(container!)
  root.render(<App />)
})
