import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'

import { useResourceDetail } from './use-resource-detail'

import fetch from 'libs/fetch'

describe('useResourceDetail', () => {
  it('should fetch information about entrypoints and middlewares', async () => {
    const { result } = renderHook(() => useResourceDetail('server-redirect@docker', 'routers'), {
      wrapper: ({ children }) => (
        <SWRConfig
          value={{
            revalidateOnFocus: false,
            fetcher: fetch,
          }}
        >
          {children}
        </SWRConfig>
      ),
    })

    await waitFor(() => {
      expect(result.current.data).not.toBeUndefined()
    })

    const { data } = result.current
    expect(data?.name).toBe('server-redirect@docker')
    expect(data?.service).toBe('api2_v2-example-beta1')
    expect(data?.status).toBe('enabled')
    expect(data?.provider).toBe('docker')
    expect(data?.rule).toBe('Host(`server`)')
    expect(data?.tls).toBeUndefined()
    expect(data?.error).toBeUndefined()
    expect(data?.middlewares?.length).toBe(1)
    expect(data?.middlewares?.[0]).toEqual({
      redirectScheme: {
        scheme: 'https',
      },
      status: 'enabled',
      usedBy: ['server-mtls@docker', 'server-redirect@docker', 'orphan-router@file'],
      name: 'redirect@file',
      type: 'redirectscheme',
      provider: 'file',
    })
    expect(data?.hasValidMiddlewares).toBeTrue()
    expect(data?.entryPointsData?.length).toBe(1)
    expect(data?.entryPointsData?.[0]).toEqual({
      address: ':80',
      transport: {
        lifeCycle: { graceTimeOut: 10000000000 },
        respondingTimeouts: { idleTimeout: 180000000000 },
      },
      forwardedHeaders: {},
      name: 'web-redirect',
    })
    expect(data?.using?.length).toBe(1)
    expect(data?.using?.[0]).toEqual('web-redirect')
  })
})
