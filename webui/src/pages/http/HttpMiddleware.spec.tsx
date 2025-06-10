import { HttpMiddlewareRender } from './HttpMiddleware'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<HttpMiddlewarePage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <HttpMiddlewareRender name="mock-middleware" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <HttpMiddlewareRender name="mock-middleware" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <HttpMiddlewareRender name="mock-middleware" data={{} as ResourceDetailDataType} error={undefined} />,
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render a simple middleware', () => {
    const mockMiddleware = {
      addPrefix: {
        prefix: '/foo',
      },
      status: 'enabled',
      usedBy: ['router-test-simple@docker'],
      name: 'middleware-simple',
      provider: 'docker',
      type: 'addprefix',
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['middleware-simple'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test-simple@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpMiddlewareRender name="mock-middleware" data={mockMiddleware as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-simple')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('addprefix')
    expect(middlewareCard.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('/foo')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test-simple@docker')
  })

  it('should render a plugin middleware', () => {
    const mockMiddleware = {
      plugin: {
        jwtAuth: {},
      },
      status: 'enabled',
      usedBy: ['router-test-plugin@docker'],
      name: 'middleware-plugin',
      provider: 'docker',
      type: 'plugin',
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['middleware-plugin'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test-plugin@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpMiddlewareRender name="mock-middleware" data={mockMiddleware as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-plugin')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('jwtAuth')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test-plugin@docker')
  })

  it('should render a complex middleware', async () => {
    const mockMiddleware = {
      name: 'middleware-complex',
      type: 'sample-middleware',
      status: 'enabled',
      provider: 'the-provider',
      usedBy: ['router-test-complex@docker'],
      redirectScheme: {
        scheme: 'redirect-scheme',
      },
      addPrefix: {
        prefix: 'add-prefix-sample',
      },
      basicAuth: {
        users: ['user1', 'user2'],
        usersFile: 'users/file',
        realm: 'realm-sample',
        removeHeader: true,
        headerField: 'basic-auth-header',
      },
      chain: {
        middlewares: ['chain-middleware-1', 'chain-middleware-2', 'chain-middleware-3'],
      },
      buffering: {
        maxRequestBodyBytes: 10000,
        memRequestBodyBytes: 10001,
        maxResponseBodyBytes: 10002,
        memResponseBodyBytes: 10003,
        retryExpression: 'buffer-retry-expression',
      },
      circuitBreaker: {
        expression: 'circuit-breaker',
      },
      compress: {},
      error: ['error-sample'],
      errors: {
        status: ['status-1', 'status-2'],
        service: 'errors-service',
        query: 'errors-query',
      },
      forwardAuth: {
        address: 'forward-auth-address',
        tls: {
          ca: 'tls-ca',
          caOptional: true,
          cert: 'tls-certificate',
          key: 'tls-key',
          insecureSkipVerify: true,
        },
        trustForwardHeader: true,
        authResponseHeaders: ['auth-response-header-1', 'auth-response-header-2'],
      },
      headers: {
        customRequestHeaders: {
          'req-header-a': 'custom-req-headers-a',
          'req-header-b': 'custom-req-headers-b',
        },
        customResponseHeaders: {
          'res-header-a': 'custom-res-headers-a',
          'res-header-b': 'custom-res-headers-b',
        },
        accessControlAllowCredentials: true,
        accessControlAllowHeaders: ['allowed-header-1', 'allowed-header-2'],
        accessControlAllowMethods: ['GET', 'POST', 'PUT'],
        accessControlAllowOrigin: 'allowed.origin',
        accessControlExposeHeaders: ['exposed-header-1', 'exposed-header-2'],
        accessControlMaxAge: 10004,
        addVaryHeader: true,
        allowedHosts: ['allowed-host-1', 'allowed-host-2'],
        hostsProxyHeaders: ['host-proxy-header-a', 'host-proxy-header-b'],
        sslRedirect: true,
        sslTemporaryRedirect: true,
        sslHost: 'ssl.host',
        sslProxyHeaders: {
          'proxy-header-a': 'ssl-proxy-header-a',
          'proxy-header-b': 'ssl-proxy-header-b',
        },
        sslForceHost: true,
        stsSeconds: 10005,
        stsIncludeSubdomains: true,
        stsPreload: true,
        forceSTSHeader: true,
        frameDeny: true,
        customFrameOptionsValue: 'custom-frame-options',
        contentTypeNosniff: true,
        browserXssFilter: true,
        customBrowserXSSValue: 'custom-xss-value',
        contentSecurityPolicy: 'content-security-policy',
        publicKey: 'public-key',
        referrerPolicy: 'referrer-policy',
        featurePolicy: 'feature-policy',
        isDevelopment: true,
      },
      ipWhiteList: {
        sourceRange: ['125.0.0.1', '125.0.0.4'],
        ipStrategy: {
          depth: 10006,
          excludedIPs: ['125.0.0.2', '125.0.0.3'],
        },
      },
      inFlightReq: {
        amount: 10007,
        sourceCriterion: {
          ipStrategy: {
            depth: 10008,
            excludedIPs: ['126.0.0.1', '126.0.0.2'],
          },
          requestHeaderName: 'inflight-req-header',
          requestHost: true,
        },
      },
      rateLimit: {
        average: 10009,
        burst: 10010,
        sourceCriterion: {
          ipStrategy: {
            depth: 10011,
            excludedIPs: ['127.0.0.1', '127.0.0.2'],
          },
          requestHeaderName: 'rate-limit-req-header',
          requestHost: true,
        },
      },
      passTLSClientCert: {
        pem: true,
        info: {
          notAfter: true,
          notBefore: true,
          sans: true,
          subject: {
            country: true,
            province: true,
            locality: true,
            organization: true,
            commonName: true,
            serialNumber: true,
            domainComponent: true,
          },
          issuer: {
            country: true,
            province: true,
            locality: true,
            organization: true,
            commonName: true,
            serialNumber: true,
            domainComponent: true,
          },
        },
      },
      redirectRegex: {
        regex: '/redirect-from-regex',
        replacement: '/redirect-to',
        permanent: true,
      },
      replacePath: {
        path: '/replace-path',
      },
      replacePathRegex: {
        regex: 'replace-path-regex',
        replacement: 'replace-path-replacement',
      },
      retry: {
        attempts: 10012,
      },
      stripPrefix: {
        prefixes: ['strip-prefix1', 'strip-prefix2'],
      },
      stripPrefixRegex: {
        regex: ['strip-prefix-regex1', 'strip-prefix-regex2'],
      },
      plugin: {
        ldapAuth: {
          source: 'plugin-ldap-source',
          baseDN: 'plugin-ldap-base-dn',
          attribute: 'plugin-ldap-attribute',
          searchFilter: 'plugin-ldap-search-filter',
          forwardUsername: true,
          forwardUsernameHeader: 'plugin-ldap-forward-username-header',
          forwardAuthorization: true,
          wwwAuthenticateHeader: true,
          wwwAuthenticateHeaderRealm: 'plugin-ldap-www-authenticate-realm',
        },
        inFlightReq: {
          amount: 10013,
          sourceCriterion: {
            ipStrategy: {
              depth: 10014,
              excludedIPs: ['128.0.0.1', '128.0.0.2'],
            },
            requestHeaderName: 'plugin-inflight-req-header',
            requestHost: true,
          },
        },
        rateLimit: {
          average: 10015,
          burst: 10016,
          sourceCriterion: {
            ipStrategy: {
              depth: 10017,
              excludedIPs: ['129.0.0.1', '129.0.0.2'],
            },
            requestHeaderName: 'plugin-rate-limit-req-header',
            requestHost: true,
          },
        },
      },
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['middleware-complex'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test-complex@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpMiddlewareRender name="mock-middleware" data={mockMiddleware as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-complex')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('the-provider')
    expect(middlewareCard.innerHTML).toContain('redirect-scheme')
    expect(middlewareCard.innerHTML).toContain('add-prefix-sample')
    expect(middlewareCard.innerHTML).toContain('buffer-retry-expression')
    expect(middlewareCard.innerHTML).toContain('circuit-breaker')
    expect(middlewareCard.innerHTML).toIncludeMultiple(['replace-path-regex', 'replace-path-replacement'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['/redirect-from-regex', '/redirect-to'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['127.0.0.1', '127.0.0.2', 'rate-limit-req-header'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['126.0.0.1', '126.0.0.2', 'inflight-req-header'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['125.0.0.1', '125.0.0.2', '125.0.0.3', '125.0.0.4'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['ssl.host', 'ssl-proxy-header-a', 'ssl-proxy-header-b'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['host-proxy-header-a', 'host-proxy-header-b'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['allowed-host-1', 'allowed-host-2'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['exposed-header-1', 'exposed-header-2'])
    expect(middlewareCard.innerHTML).toContain('allowed.origin')
    expect(middlewareCard.innerHTML).toContain('custom-frame-options')
    expect(middlewareCard.innerHTML).toContain('content-security-policy')
    expect(middlewareCard.innerHTML).toContain('public-key')
    expect(middlewareCard.innerHTML).toContain('referrer-policy')
    expect(middlewareCard.innerHTML).toContain('feature-policy')
    expect(middlewareCard.innerHTML).toIncludeMultiple(['GET', 'POST', 'PUT'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['allowed-header-1', 'allowed-header-2'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['custom-res-headers-a', 'custom-res-headers-b'])
    expect(middlewareCard.innerHTML).toIncludeMultiple(['custom-req-headers-a', 'custom-req-headers-b'])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'forward-auth-address',
      'auth-response-header-1',
      'auth-response-header-2',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'error-sample',
      'status-1',
      'status-2',
      'errors-service',
      'errors-query',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'chain-middleware-1',
      'chain-middleware-2',
      'chain-middleware-3',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'user1',
      'user2',
      'users/file',
      'realm-sample',
      'basic-auth-header',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'strip-prefix1',
      'strip-prefix2',
      'strip-prefix-regex1',
      'strip-prefix-regex2',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      '10000',
      '10001',
      '10002',
      '10003',
      '10004',
      '10005',
      '10006',
      '10007',
      '10008',
      '10009',
      '10010',
      '10011',
      '10012',
    ])
    expect(middlewareCard.innerHTML).toIncludeMultiple([
      'plugin-ldap-source',
      'plugin-ldap-base-dn',
      'plugin-ldap-attribute',
      'plugin-ldap-search-filter',
      'plugin-ldap-forward-username-header',
      'plugin-ldap-www-authenticate-realm',
      'plugin-inflight-req-header',
      'plugin-rate-limit-req-header',
      '10013',
      '10014',
      '10015',
      '10016',
      '10017',
    ])

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test-complex@docker')
  })

  it('should render a plugin middleware with no type', async () => {
    const mockMiddleware = {
      plugin: {
        jwtAuth: {
          child: {},
          sibling: {
            negativeGrandChild: false,
            positiveGrandChild: true,
          },
          stringChild: '123',
          arrayChild: [1, 2, 3],
        },
      },
      status: 'enabled',
      name: 'middleware-plugin-no-type',
      provider: 'docker',
      routers: [],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpMiddlewareRender name="mock-middleware" data={mockMiddleware as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-plugin-no-type')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('jwtAuth &gt; child')
    expect(middlewareCard.innerHTML).toContain('jwtAuth &gt; sibling &gt; negative Grand Child')
    expect(middlewareCard.innerHTML).toContain('jwtAuth &gt; sibling &gt; positive Grand Child')
    expect(middlewareCard.innerHTML).toContain('jwtAuth &gt; string Child')
    expect(middlewareCard.innerHTML).toContain('jwtAuth &gt; array Child')

    const childSpans = Array.from(middlewareCard.querySelectorAll('span')).filter((span) =>
      ['0', '1', '2', '3', '123'].includes(span.innerHTML),
    )
    expect(childSpans.length).toBe(7)
  })
})
