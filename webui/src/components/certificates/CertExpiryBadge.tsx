import { Badge } from '@traefiklabs/faency'

type ExpiryStatus = {
  variant: 'red' | 'orange' | 'green'
  label: string
}

export const getCertExpiryStatus = (daysLeft: number): ExpiryStatus => {
  if (daysLeft < 0) return { variant: 'red', label: 'EXPIRED' }
  if (daysLeft < 14) return { variant: 'orange', label: 'Expiring Soon' }
  return { variant: 'green', label: 'Valid' }
}

type CertExpiryBadgeProps = {
  daysLeft: number
  size?: 'small' | 'large'
}

const CertExpiryBadge = ({ daysLeft, size = 'large' }: CertExpiryBadgeProps) => {
  const { variant, label } = getCertExpiryStatus(daysLeft)

  return (
    <Badge size={size} variant={variant}>
      {daysLeft < 0 ? 'EXPIRED' : label === 'Valid' ? `${daysLeft} days` : `${daysLeft} days`}
    </Badge>
  )
}

export default CertExpiryBadge
