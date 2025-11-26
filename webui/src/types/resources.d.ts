declare namespace Service {
  type WeightedService = {
    name: string
    weight: number
  }

  type Mirror = {
    name: string
    percent: number
  }

  type Details = {
    name: string
    status: 'enabled' | 'disabled' | 'warning'
    provider: string
    type: string
    usedBy?: string[]
    routers?: Router[]
    serverStatus?: {
      [server: string]: string
    }
    mirroring?: {
      service: string
      mirrors?: Service.Mirror[]
    }
    loadBalancer?: {
      servers?: { url: string }[]
      passHostHeader?: boolean
      terminationDelay?: number
      healthCheck?: {
        scheme: string
        path: string
        hostname: string
        headers?: {
          [header: string]: string
        }
        port?: number
        send?: string
        expect?: string
        interval?: string
        unhealthyInterval?: string
        timeout?: string
      }
    }
    weighted?: {
      services?: Service.WeightedService[]
    }
  }
}

declare namespace Middleware {
  type MiddlewareProps = {
    [prop: string]: ValuesMapType
  }

  type Details = {
    name: string
    status: 'enabled' | 'disabled' | 'warning'
    provider: string
    type?: string
    plugin?: Record<string, unknown>
    error?: string[]
    routers?: string[]
    usedBy?: string[]
  } & MiddlewareProps
}
