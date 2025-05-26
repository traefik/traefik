import { useMemo } from 'react'

type ProviderIconProps = {
  name: string
  size?: number
}

const getProviderName = (providerName: string) => {
  if (!providerName || typeof providerName !== 'string') return 'internal'

  const name = providerName.toLowerCase()

  if (name.startsWith('plugin-')) {
    return 'plugin'
  }
  if (name.startsWith('kubernetes')) {
    return 'kubernetes'
  }
  if (name.startsWith('consul-')) {
    return 'consul'
  }
  if (name.startsWith('consulcatalog-')) {
    return 'consulcatalog'
  }
  if (name.startsWith('nomad-')) {
    return 'nomad'
  }

  return name
}

export const ProviderIcon = ({ name, size = 32 }: ProviderIconProps) => {
  const src = useMemo(() => `${import.meta.env.BASE_URL || '/'}img/providers/${getProviderName(name)}.svg`, [name])

  return (
    <img
      alt={name}
      src={src}
      width={size}
      height={size}
      style={{ backgroundColor: 'var(--colors-primary)', borderRadius: '50%' }}
    />
  )
}
