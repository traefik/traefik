import { Box } from '@traefiklabs/faency'
import { HTMLAttributes, useMemo } from 'react'

import Consul from 'components/icons/providers/Consul'
import Docker from 'components/icons/providers/Docker'
import ECS from 'components/icons/providers/ECS'
import Etcd from 'components/icons/providers/Etcd'
import File from 'components/icons/providers/File'
import Http from 'components/icons/providers/Http'
import Hub from 'components/icons/providers/Hub'
import Internal from 'components/icons/providers/Internal'
import Knative from 'components/icons/providers/Knative'
import Kubernetes from 'components/icons/providers/Kubernetes'
import Nomad from 'components/icons/providers/Nomad'
import Plugin from 'components/icons/providers/Plugin'
import Redis from 'components/icons/providers/Redis'
import Zookeeper from 'components/icons/providers/Zookeeper'
import Tooltip from 'components/Tooltip'

export type ProviderIconProps = HTMLAttributes<SVGElement> & {
  height?: number | string
  width?: number | string
}

export default function ProviderIcon({ name, size = 20 }: { name: string; size?: number }) {
  const Icon = useMemo(() => {
    if (!name || typeof name !== 'string') return Internal

    const nameLowerCase = name.toLowerCase()

    if (['consul', 'consul-', 'consulcatalog-'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Consul
    }
    if (['docker', 'swarm'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Docker
    }
    if (['ecs'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return ECS
    }
    if (['etcd'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Etcd
    }
    if (['file'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return File
    }
    if (['http'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Http
    }
    if (['hub'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Hub
    }
    if (['kubernetes'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Kubernetes
    }
    if (['knative'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Knative
    }
    if (['nomad', 'nomad-'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Nomad
    }
    if (['plugin', 'plugin-'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Plugin
    }
    if (['redis'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Redis
    }
    if (['zookeeper'].some((prefix) => nameLowerCase.startsWith(prefix))) {
      return Zookeeper
    }
    return Internal
  }, [name])

  return (
    <Icon
      height={size}
      width={size}
      style={{ backgroundColor: 'var(--colors-primary)', borderRadius: '50%', color: 'var(--colors-01dp)' }}
    />
  )
}

export const ProviderIconWithTooltip = ({ provider, size = 20 }) => {
  return (
    <Tooltip label={provider}>
      <Box css={{ width: size, height: size }}>
        <ProviderIcon name={provider} size={size} />
      </Box>
    </Tooltip>
  )
}
