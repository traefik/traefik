import { Box, Card, Flex, H3, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Chart as ChartJs, ArcElement, Tooltip } from 'chart.js'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { Doughnut } from 'react-chartjs-2'
import { FaArrowRightLong } from 'react-icons/fa6'
import { Link as RouterLink, useNavigate } from 'react-router-dom'

import Status, { colorByStatus, StatusType } from './Status'

import { capitalizeFirstLetter } from 'utils/string'

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
  linkTo: string
}

export type DataType = {
  datasets: {
    backgroundColor: string[]
    data: (string | number)[]
  }[]
  labels?: string[]
}

const getPercent = (total: number, value: number) => (total > 0 ? ((value * 100) / total).toFixed(0) : 0)

const STATS_ATTRIBUTES: { status: StatusType; label: string }[] = [
  {
    status: 'enabled',
    label: 'success',
  },
  {
    status: 'warning',
    label: 'warnings',
  },
  {
    status: 'disabled',
    label: 'errors',
  },
]

const CustomLegend = ({
  status,
  label,
  count,
  total,
  linkTo,
}: {
  status: StatusType
  label: string
  count: number
  total: number
  linkTo: string
}) => {
  return (
    <Link to={`${linkTo}?status=${status}`}>
      <Flex css={{ alignItems: 'center', p: '$2' }}>
        <Status status={status} />
        <Flex css={{ flexDirection: 'column', flex: 1 }}>
          <Text css={{ fontWeight: 600 }}>{capitalizeFirstLetter(label)}</Text>
          <Text size={1} css={{ color: 'hsl(0, 0%, 56%)' }} data-testid={`${label}-pc`}>
            {getPercent(total, count)}%
          </Text>
        </Flex>
        <Text size={5} css={{ fontWeight: 700 }} data-testid={`${label}-count`}>
          {count}
        </Text>
      </Flex>
    </Link>
  )
}

const TraefikResourceStatsCard = ({ title, errors, total, warnings, linkTo }: TraefikResourceStatsCardProps) => {
  const navigate = useNavigate()

  const defaultData = {
    datasets: [
      {
        backgroundColor: [colorByStatus.enabled],
        data: [1],
      },
    ],
  }
  const [data, setData] = useState<DataType>(defaultData)

  const counts = useMemo(
    () => ({
      success: total - (errors + warnings),
      warnings,
      errors,
    }),
    [errors, total, warnings],
  )

  useEffect(() => {
    if (counts.success + counts.warnings + counts.errors === 0) {
      setData(defaultData)
      return
    }

    const newData = {
      datasets: [
        {
          backgroundColor: [colorByStatus.enabled, colorByStatus.warning, colorByStatus.error],
          data: [counts.success, counts.warnings, counts.errors],
        },
      ],
      labels: ['Success', 'Warnings', 'Errors'],
    }

    setData(newData)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [errors, warnings, total, counts])

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
    onClick: (_, activeEl) => {
      navigate(`${linkTo}?status=${STATS_ATTRIBUTES[activeEl[0].index].status}`)
    },
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
                <Flex align="center" gap={1} css={{ color: '$primary' }}>
                  <Text css={{ fontWeight: 500, color: '$primary' }}>Explore</Text>
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
          {STATS_ATTRIBUTES.map((i) => (
            <CustomLegend key={`${title}-${i.label}`} {...i} count={counts[i.label]} total={total} linkTo={linkTo} />
          ))}
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
