import { Box, Text } from '@traefiklabs/faency'
import { Badge } from 'components/FaencyOverrides'
import Tooltip from 'components/Tooltip'
import { MiddlewareProps } from 'hooks/use-resource-detail'

function capitalize(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1)
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function quote(value: any): any {
  if (typeof value === 'string') {
    return `"${value}"`
  }

  return value
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function quoteArray(values: any[]): any {
  return values.map(quote)
}

type JSONObject = {
  [x: string]: any // eslint-disable-line @typescript-eslint/no-explicit-any
}
type ValuesMapType = {
  [key: string]: string | number | JSONObject
}
const renderFeatureValues = (valuesMap: ValuesMapType): string => {
  return Object.entries(valuesMap)
    .map(([name, value]) => {
      const capitalizedName = capitalize(name)
      if (typeof value === 'string') {
        return [capitalizedName, `"${value}"`].join('=')
      }

      if (value instanceof Array) {
        return [capitalizedName, quoteArray(value).join(', ')].join('=')
      }

      if (typeof value === 'object') {
        return [capitalizedName, `{${renderFeatureValues(value)}}`].join('=')
      }

      return [capitalizedName, value].join('=')
    })
    .join(', ')
}

const FeatureMiddleware = ({ middleware }: { middleware: MiddlewareProps }) => {
  const [name, value] = Object.entries(middleware)[0]
  const content = `${capitalize(name)}: ${renderFeatureValues(value)}`

  return (
    <Tooltip label={content} action="copy">
      <Badge variant="deepBlue" css={{ mr: '$2', mt: '$2' }}>
        {content}
      </Badge>
    </Tooltip>
  )
}

type AdditionalFeaturesProps = {
  middlewares?: MiddlewareProps[]
  uid: string
}

const AdditionalFeatures = ({ middlewares, uid }: AdditionalFeaturesProps) => {
  return middlewares?.length ? (
    <Box css={{ mt: '-$2' }}>
      {middlewares.map((m, idx) => (
        <FeatureMiddleware key={`${uid}-${idx}`} middleware={m} />
      ))}
    </Box>
  ) : (
    <Text css={{ fontStyle: 'italic', color: 'hsl(0, 0%, 56%)' }}>No additional features</Text>
  )
}

export default AdditionalFeatures
