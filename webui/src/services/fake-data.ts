export const fakeData = {
  'provider0': {
    'backends': {
      'provider0Backend0': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider0Backend1': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider0Backend2': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider0Backend3': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      }
    },
    'frontends': {
      'Frontend0': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider0Backend0',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend1': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider0Backend1',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend2': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider0Backend2',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend3': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider0Backend3',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      }
    }
  },
  'provider1': {
    'backends': {
      'provider1Backend0': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider1Backend1': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider1Backend2': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider1Backend3': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      }
    },
    'frontends': {
      'Frontend0': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider1Backend0',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend1': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider1Backend1',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend2': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider1Backend2',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend3': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider1Backend3',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      }
    }
  },
  'provider2': {
    'backends': {
      'provider2Backend0': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider2Backend1': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider2Backend2': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider2Backend3': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      }
    },
    'frontends': {
      'Frontend0': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider2Backend0',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend1': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider2Backend1',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend2': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider2Backend2',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend3': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider2Backend3',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      }
    }
  },
  'provider3': {
    'backends': {
      'provider3Backend0': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider3Backend1': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider3Backend2': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      },
      'provider3Backend3': {
        'servers': {
          'Server0': {
            'url': 'foobar',
            'weight': 42
          },
          'Server1': {
            'url': 'foobar',
            'weight': 42
          },
          'Server2': {
            'url': 'foobar',
            'weight': 42
          },
          'Server3': {
            'url': 'foobar',
            'weight': 42
          }
        },
        'circuitBreaker': {
          'expression': 'foobar'
        },
        'loadBalancer': {
          'method': 'foobar',
          'sticky': true,
          'stickiness': {
            'cookieName': 'foobar'
          }
        },
        'maxConn': {
          'amount': 42,
          'extractorFunc': 'foobar'
        },
        'healthCheck': {
          'path': 'foobar',
          'port': 42,
          'interval': 'foobar'
        }
      }
    },
    'frontends': {
      'Frontend0': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider3Backend0',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend1': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider3Backend1',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend2': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider3Backend2',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      },
      'Frontend3': {
        'entryPoints': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'backend': 'provider3Backend3',
        'routes': {
          'Route0': {
            'rule': 'foobar'
          },
          'Route1': {
            'rule': 'foobar'
          },
          'Route2': {
            'rule': 'foobar'
          },
          'Route3': {
            'rule': 'foobar'
          }
        },
        'passHostHeader': true,
        'passTLSCert': true,
        'priority': 42,
        'basicAuth': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'whitelistSourceRange': [
          'foobar',
          'foobar',
          'foobar'
        ],
        'headers': {
          'customRequestHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'customResponseHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'allowedHosts': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'hostsProxyHeaders': [
            'foobar',
            'foobar',
            'foobar'
          ],
          'sslRedirect': true,
          'sslTemporaryRedirect': true,
          'sslHost': 'foobar',
          'sslProxyHeaders': {
            'name0': 'foobar',
            'name1': 'foobar',
            'name2': 'foobar',
            'name3': 'foobar'
          },
          'stsSeconds': 42,
          'stsIncludeSubdomains': true,
          'stsPreload': true,
          'forceSTSHeader': true,
          'frameDeny': true,
          'customFrameOptionsValue': 'foobar',
          'contentTypeNosniff': true,
          'browserXssFilter': true,
          'contentSecurityPolicy': 'foobar',
          'publicKey': 'foobar',
          'referrerPolicy': 'foobar',
          'isDevelopment': true
        },
        'errors': {
          'ErrorPage0': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage1': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage2': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          },
          'ErrorPage3': {
            'status': [
              'foobar',
              'foobar',
              'foobar'
            ],
            'backend': 'foobar',
            'query': 'foobar'
          }
        }
      }
    }
  }
};
