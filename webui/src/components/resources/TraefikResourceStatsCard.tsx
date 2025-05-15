import { Box, Card, Flex, H3, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Chart as ChartJs, ArcElement, Tooltip } from 'chart.js'
import { ReactNode, useCallback, useEffect, useState } from 'react'
import { Doughnut } from 'react-chartjs-2'
import { FaArrowRightLong } from 'react-icons/fa6'
import { Link as RouterLink } from 'react-router-dom'

import Status, { colorByStatus } from './Status'

ChartJs.register(ArcElement, Tooltip)

const Link = styled(RouterLink, {
  textDecoration: 'none',

  '&:hover': {
    textDecoration: 'none',
  },
})

type StatsCardType = {
  children: ReactNode
}

const StatsCard = ({ children, ...props }: StatsCardType) => (
  <Card
    css={{
      display: 'flex',
      flexDirection: 'column',
      padding: '16px',
      overflow: 'hidden',
    }}
    {...props}
  >
    {children}
  </Card>
)

export type TraefikResourceStatsType = {
  title?: string
  errors: number
  total: number
  warnings: number
}

export type TraefikResourceStatsCardProps = TraefikResourceStatsType & {
  linkTo?: string
}

export type DataType = {
  datasets: {
    backgroundColor: string[]
    data: (string | number)[]
  }[]
  labels?: string[]
}

const TraefikResourceStatsCard = ({ title, errors, total, warnings, linkTo }: TraefikResourceStatsCardProps) => {
  const defaultData = {
    datasets: [
      {
        backgroundColor: [colorByStatus.enabled],
        data: [1],
      },
    ],
  }
  const [data, setData] = useState<DataType>(defaultData)
  const getPercent = useCallback((value: number) => (total > 0 ? (value * 100) / total : 0), [total])

  const getSuccess = useCallback(
    (inPercent = false): number | string => {
      const num = total - (errors + warnings)

      if (inPercent) {
        return getPercent(num).toFixed(0)
      } else {
        return num
      }
    },
    [errors, total, warnings, getPercent],
  )

  const getWarnings = useCallback(
    (inPercent = false): number | string => {
      const num = warnings

      if (inPercent) {
        return getPercent(num).toFixed(0)
      } else {
        return num
      }
    },
    [warnings, getPercent],
  )

  const getErrors = useCallback(
    (inPercent = false): number | string => {
      const num = errors

      if (inPercent) {
        return getPercent(num).toFixed(0)
      } else {
        return num
      }
    },
    [errors, getPercent],
  )

  useEffect(() => {
    const successCount = getSuccess() as number
    const warningsCount = getWarnings() as number
    const errorsCount = getErrors() as number

    if (successCount + warningsCount + errorsCount === 0) {
      setData(defaultData)
      return
    }

    const newData = {
      datasets: [
        {
          backgroundColor: [colorByStatus.enabled, colorByStatus.warning, colorByStatus.error],
          data: [successCount, warningsCount, errorsCount],
        },
      ],
      labels: ['Success', 'Warnings', 'Errors'],
    }

    setData(newData)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [errors, warnings, total, getSuccess, getWarnings, getErrors])

  const options = {
    animation: {
      duration: 1000,
    },
    plugins: {
      legend: {
        display: false,
      },
    },
    tooltips: {
      enabled: true,
    },
    maintainAspectRatio: false,
  }

  if (!errors && !total && !warnings) return null

  return (
    <StatsCard data-testid="card">
      {title && (
        <Flex css={{ pb: '$3', mb: '$2' }}>
          {title && (
            <Flex align="center" justify="space-between" css={{ flex: '1' }}>
              <H3 css={{ fontSize: '$6' }}>{title}</H3>

              <Link to={linkTo as string}>
                <Flex align="center" gap={1} css={{ color: '$linkBlue' }}>
                  <Text css={{ fontWeight: 500, color: '$linkBlue' }}>Explore</Text>
                  <FaArrowRightLong />
                </Flex>
              </Link>
            </Flex>
          )}
        </Flex>
      )}
      <Flex css={{ flex: '1' }}>
        <Box css={{ width: '50%' }}>
          <Doughnut data={data} options={options} />
        </Box>
        <Box css={{ width: '50%' }}>
          <Flex css={{ alignItems: 'center', p: '$2' }}>
            <Status status="enabled" />
            <Flex css={{ flexDirection: 'column', flex: 1 }}>
              <Text css={{ fontWeight: 600 }}>Success</Text>
              <Text size={1} css={{ color: 'hsl(0, 0%, 56%)' }} data-testid="success-pc">
                {getSuccess(true)}%
              </Text>
            </Flex>
            <Text size={5} css={{ fontWeight: 700 }} data-testid="success-count">
              {getSuccess()}
            </Text>
          </Flex>
          <Flex css={{ alignItems: 'center', p: '$2' }}>
            <Status status="warning" />
            <Flex css={{ flexDirection: 'column', flex: 1 }}>
              <Text css={{ fontWeight: 600 }}>Warnings</Text>
              <Text size={1} css={{ color: 'hsl(0, 0%, 56%)' }} data-testid="warnings-pc">
                {getWarnings(true)}%
              </Text>
            </Flex>
            <Text size={5} css={{ fontWeight: 700 }} data-testid="warnings-count">
              {getWarnings()}
            </Text>
          </Flex>
          <Flex css={{ alignItems: 'center', p: '$2' }}>
            <Status status="disabled" />
            <Flex css={{ flexDirection: 'column', flex: 1 }}>
              <Text css={{ fontWeight: 600 }}>Errors</Text>
              <Text size={1} css={{ color: 'hsl(0, 0%, 56%)' }} data-testid="errors-pc">
                {getErrors(true)}%
              </Text>
            </Flex>
            <Text size={5} css={{ fontWeight: 700 }} data-testid="errors-count">
              {getErrors()}
            </Text>
          </Flex>
        </Box>
      </Flex>
    </StatsCard>
  )
}

export const StatsCardSkeleton = () => {
  return (
    <StatsCard>
      <Flex gap={2}>
        <Skeleton css={{ width: '80%', height: 150 }} />
        <Flex direction="column" gap={2} css={{ flex: 1 }}>
          <Skeleton />
          <Skeleton />
          <Skeleton />
        </Flex>
      </Flex>
    </StatsCard>
  )
}

export default TraefikResourceStatsCard
