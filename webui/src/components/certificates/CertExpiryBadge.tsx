import { Badge } from '@traefik-labs/faency'

type ExpiryStatus = {
  variant: 'red' | 'orange' | 'green'
  label: string
}

// Ratio of the remaining lifetime of a certificate before it is considered expiring
const expiringRatio = 1 / 3;

export const getCertExpiryStatus = (daysLeft: number, daysLifetime: number): ExpiryStatus => {
  if (daysLeft < 0) return { variant: 'red', label: 'EXPIRED' }
  if (daysLeft / daysLifetime < expiringRatio) return { variant: 'orange', label: 'Expiring Soon' }
  return { variant: 'green', label: 'Valid' }
}

type CertExpiryBadgeProps = {
  daysLeft: number
  daysLifetime: number
  size?: 'small' | 'large'
}

const CertExpiryBadge = ({ daysLeft, daysLifetime, size = 'large' }: CertExpiryBadgeProps) => {
  const { variant } = getCertExpiryStatus(daysLeft, daysLifetime)

  return (
    <Badge size={size} variant={variant}>
      {daysLeft < 0 ? 'EXPIRED' : `${daysLeft} days`}
    </Badge>
  )
}

export default CertExpiryBadge
