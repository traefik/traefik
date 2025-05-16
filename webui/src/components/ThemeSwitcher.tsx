import { AccessibleIcon, Button } from '@traefiklabs/faency'
import { FiMoon, FiSun } from 'react-icons/fi'

import { AutoThemeIcon } from 'components/icons/AutoThemeIcon'
import { useTheme } from 'hooks/use-theme'

export default function ThemeSwitcher() {
  const { selectedTheme, setTheme } = useTheme()

  return (
    <Button
      ghost
      css={{ px: '$2', color: '$buttonSecondaryText' }}
      onClick={setTheme}
      type="button"
      data-testid="theme-switcher"
    >
      <AccessibleIcon label="toggle theme">
        {selectedTheme === 'dark' ? (
          <FiMoon size={20} />
        ) : selectedTheme === 'light' ? (
          <FiSun size={20} />
        ) : (
          <AutoThemeIcon />
        )}
      </AccessibleIcon>
    </Button>
  )
}
