import { AccessibleIcon, Button } from '@traefiklabs/faency'
import { CgDarkMode } from 'react-icons/cg'
import { FiMoon, FiSun } from 'react-icons/fi'

import { useTheme } from 'hooks/use-theme'

export default function ThemeSwitcher() {
  const { selectedTheme, setTheme } = useTheme()

  return (
    <Button ghost css={{ px: '$2' }} onClick={setTheme} type="button" data-testid="theme-switcher">
      <AccessibleIcon label="toggle theme">
        {selectedTheme === 'dark' ? (
          <FiMoon size={20} />
        ) : selectedTheme === 'light' ? (
          <FiSun size={20} />
        ) : (
          <CgDarkMode size={22} />
        )}
      </AccessibleIcon>
    </Button>
  )
}
