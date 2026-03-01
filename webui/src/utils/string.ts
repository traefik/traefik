export const capitalizeFirstLetter = (string: unknown): string | null => {
  if (!string) return null

  return string?.toString()?.charAt(0)?.toUpperCase() + string?.toString()?.slice(1)
}
