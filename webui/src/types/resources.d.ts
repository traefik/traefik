declare namespace Resource {
  type Status = 'info' | 'success' | 'warning' | 'error' | 'enabled' | 'disabled' | 'loading'

  type DetailsData = Router.DetailsData & Service.Details & Middleware.DetailsData
}

declare namespace Entrypoint {
  type Details = {
    name: string
    address: string
    message?: string
  }
}

declare namespace Router {
  type TlsDomain = {
    main: string
    sans: string[]
  }

  type TLS = {
    options?: string
    certResolver: string
    domains: TlsDomain[]
    passthrough?: boolean
  }

  type Details = {
    name: string
    service?: string
    status: 'enabled' | 'disabled' | 'warning'
    rule?: string
    priority?: number
    provider: string
    tls?: {
      options: string
      certResolver: string
      domains: TlsDomain[]
      passthrough: boolean
    }
    error?: string[]
    entryPoints?: string[]
    message?: string
  }

  type DetailsData = Details & {
    middlewares?: Middleware.Details[]
    hasValidMiddlewares?: boolean
    entryPointsData?: Entrypoint.Details[]
    using?: string[]
  }
}

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
      mirrors?: Mirror[]
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
      services?: WeightedService[]
    }
  }
}

declare namespace Middleware {
  type Props = {
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
  } & Props

  type DetailsData = Details & {
    routers?: Router.Details[]
  }
}

declare namespace Certificate {
  /** Raw API response shape */
  type Raw = {
    name?: string
    commonName: string
    sans: string[]
    issuer?: string
    issuerOrg?: string
    issuerCN?: string
    issuerCountry?: string
    organization?: string
    country?: string
    subject?: string
    serialNumber?: string
    notBefore: string
    notAfter: string
    version?: string
    keyType?: string
    keySize?: number
    signatureAlgorithm?: string
    certFingerprint?: string
    publicKeyFingerprint?: string
    status?: 'enabled' | 'disabled' | 'warning'
    resolver?: string
    usedBy?: string[]
  }

  /** Enriched certificate with computed fields */
  type Info = Raw & {
    daysLeft: number
  }
}
