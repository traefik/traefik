import { buildCertKey } from './use-certificates'

describe('buildCertKey', () => {
  it('should build a certKey from a single domain', () => {
    const certKey = buildCertKey('example.com', [])
    expect(certKey).toBe(btoa('example.com'))
  })

  it('should build a certKey from multiple domains', () => {
    const certKey = buildCertKey('example.com', ['www.example.com', 'api.example.com'])
    // Should deduplicate, lowercase, and sort
    expect(certKey).toBe(btoa('api.example.com,example.com,www.example.com'))
  })

  it('should handle duplicate domains', () => {
    const certKey = buildCertKey('example.com', ['example.com', 'Example.Com', 'EXAMPLE.COM'])
    // Should deduplicate and lowercase
    expect(certKey).toBe(btoa('example.com'))
  })

  it('should handle IP addresses', () => {
    const certKey = buildCertKey('127.0.0.1', ['::1', 'example.com'])
    expect(certKey).toBe(btoa('127.0.0.1,::1,example.com'))
  })

  it('should sort domains alphabetically', () => {
    const certKey = buildCertKey('zebra.com', ['apple.com', 'banana.com'])
    expect(certKey).toBe(btoa('apple.com,banana.com,zebra.com'))
  })
})
