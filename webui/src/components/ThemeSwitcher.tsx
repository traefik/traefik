import { AccessibleIcon } from '@traefiklabs/faency'
import { useDarkMode } from 'hooks/use-dark-mode'
import { FiMoon, FiSun } from 'react-icons/fi'

import { Button } from './FaencyOverrides'

export const ThemeSwitcher = () => {
  const { isDarkMode, toggle } = useDarkMode({ initializeWithValue: false })

  return (
    <Button ghost css={{ px: '$2' }} onClick={toggle} type="button" data-testid="theme-switcher">
      <AccessibleIcon label="toggle theme">{isDarkMode ? <FiMoon size={20} /> : <FiSun size={20} />}</AccessibleIcon>
    </Button>
  )
}

export default ThemeSwitcher
