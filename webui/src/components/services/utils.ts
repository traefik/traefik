export const getProviderFromName = (serviceName: string, defaultProvider: string): string => {
  const [, provider] = serviceName.split('@')
  return provider || defaultProvider
}
