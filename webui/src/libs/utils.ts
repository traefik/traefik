const getBasePath = () => {
  const { VITE_APP_BASE_API_URL } = import.meta.env
  const basePath = window.APIUrl || VITE_APP_BASE_API_URL || ''

  return basePath.endsWith('/') ? basePath.slice(0, -1) : basePath
}

export const BASE_PATH = getBasePath()
