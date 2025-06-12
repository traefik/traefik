import { useMemo } from 'react'
import { useLocalStorage } from 'usehooks-ts'

const SYSTEM = 'system'
const DARK = 'dark'
const LIGHT = 'light'

type ThemeOptions = 'system' | 'dark' | 'light'
const THEME_OPTIONS: ThemeOptions[] = [SYSTEM, DARK, LIGHT]

type UseThemeRes = {
  selectedTheme: ThemeOptions
  appliedTheme: ThemeOptions
  setTheme: () => void
}

export const useTheme = (): UseThemeRes => {
  const [selectedTheme, setSelectedTheme] = useLocalStorage<ThemeOptions>('selected-theme', SYSTEM)
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  const appliedTheme = useMemo(() => {
    if (selectedTheme === SYSTEM) {
      if (prefersDark) return DARK
      return LIGHT
    }

    return selectedTheme
  }, [selectedTheme, prefersDark])

  return {
    selectedTheme,
    appliedTheme,
    setTheme: () => {
      setSelectedTheme((curr: ThemeOptions): ThemeOptions => {
        const currIdx = THEME_OPTIONS.indexOf(curr)
        const nextIdx = currIdx + 1
        if (nextIdx === THEME_OPTIONS.length) return SYSTEM

        return THEME_OPTIONS[nextIdx]
      })
    },
  }
}

export const useIsDarkMode = () => {
  const { appliedTheme } = useTheme()

  return appliedTheme === DARK
}
