import { Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'

import { BooleanState, ItemBlock } from './DetailSections'
import GenericTable from './GenericTable'
import IpStrategyTable, { IpStrategy } from './IpStrategyTable'

import Tooltip from 'components/Tooltip'

type RenderUnknownPropProps = {
  name: string
  prop?: unknown
  removeTitlePrefix?: string
}

export const RenderUnknownProp = ({ name, prop, removeTitlePrefix }: RenderUnknownPropProps) => {
  const wrap = (children: ReactNode, altName?: string, key?: string) => (
    <ItemBlock key={key} title={altName || name}>
      {children}
    </ItemBlock>
  )
  try {
    if (typeof prop !== 'undefined') {
      if (typeof prop === 'boolean') {
        return wrap(<BooleanState enabled={prop} />)
      }

      if (typeof prop === 'string' && ['true', 'false'].includes((prop as string).toLowerCase())) {
        return wrap(<BooleanState enabled={prop === 'true'} />)
      }

      if (['string', 'number'].includes(typeof prop)) {
        return wrap(
          <Tooltip label={prop as string} action="copy">
            <Text css={{ overflowWrap: 'break-word' }}>{prop as string}</Text>
          </Tooltip>,
        )
      }

      if (JSON.stringify(prop) === '{}') {
        return wrap(<BooleanState enabled />)
      }

      if (prop instanceof Array) {
        return wrap(
          <GenericTable items={prop.map((p) => (['number', 'string'].includes(typeof p) ? p : JSON.stringify(p)))} />,
        )
      }

      if (prop?.constructor === Object) {
        return (
          <>
            {Object.entries(prop).map(([childName, childProp]) => {
              const spacedChildName = childName.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
              let title = `${name} > ${spacedChildName}`
              if (removeTitlePrefix) {
                title = title.replace(new RegExp(`^${removeTitlePrefix} > `, 'i'), '')
              }

              switch (childName) {
                case 'ipStrategy':
                  return wrap(<IpStrategyTable ipStrategy={childProp as IpStrategy} />, title, title)
                case 'statusRewrites':
                  return wrap(
                    <GenericTable items={Object.entries(childProp).map((x) => `${x[0]} → ${x[1]}`)} />,
                    title,
                    title,
                  )
                default:
                  return <RenderUnknownProp key={title} name={title} prop={childProp} />
              }
            })}
          </>
        )
      }
    }
  } catch (error) {
    console.log('Unable to render plugin property:', { name, prop }, { error })
  }

  return null
}
