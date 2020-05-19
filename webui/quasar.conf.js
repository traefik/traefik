// Configuration for your app
// https://quasar.dev/quasar-cli/quasar-conf-js

module.exports = function (ctx) {
  return {
    // app boot file (/src/boot)
    // --> boot files are part of "main.js"
    boot: [
      '_globals',
      'api',
      '_hacks',
      '_init'
    ],

    css: [
      'sass/app.scss'
    ],

    extras: [
      // 'ionicons-v4',
      // 'mdi-v3',
      // 'fontawesome-v5',
      'eva-icons',
      // 'themify',
      // 'roboto-font-latin-ext', // this or either 'roboto-font', NEVER both!

      'roboto-font', // optional, you are not bound to it
      'material-icons' // optional, you are not bound to it
    ],

    framework: {
      // iconSet: 'ionicons-v4',
      // lang: 'de', // Quasar language

      // all: true, // --- includes everything; for dev only!

      components: [
        'QLayout',
        'QHeader',
        'QFooter',
        'QDrawer',
        'QPageContainer',
        'QPage',
        'QPageSticky',
        'QPageScroller',
        'QToolbar',
        'QSpace',
        'QToolbarTitle',
        'QTooltip',
        'QBtn',
        'QIcon',
        'QList',
        'QItem',
        'QExpansionItem',
        'QItemSection',
        'QItemLabel',
        'QTabs',
        'QTab',
        'QRouteTab',
        'QAvatar',
        'QSeparator',
        'QScrollArea',
        'QImg',
        'QBadge',
        'QCard',
        'QCardSection',
        'QCardActions',
        'QBreadcrumbs',
        'QBreadcrumbsEl',
        'QInput',
        'QToggle',
        'QForm',
        'QField',
        'QSelect',
        'QCheckbox',
        'QRadio',
        'QMenu',
        'QAjaxBar',
        'QTable',
        'QTh',
        'QTr',
        'QTd',
        'QFab',
        'QFabAction',
        'QDialog',
        'QUploader',
        'QTree',
        'QChip',
        'QBtnToggle'
      ],

      directives: [
        'ClosePopup',
        'Ripple'
      ],

      // Quasar plugins
      plugins: [
        'Notify',
        'Dialog',
        'LoadingBar'
      ],

      config: {
        notify: { /* Notify defaults */ },
        loadingBar: {
          position: 'top',
          color: 'accent',
          size: '2px'
        }
      }
    },

    supportIE: false,

    build: {
      publicPath: process.env.APP_PUBLIC_PATH || '',
      env: process.env.APP_ENV === 'development'
        ? { // staging:
          APP_ENV: JSON.stringify(process.env.APP_ENV),
          APP_API: JSON.stringify(process.env.APP_API || '/api'),
          PLATFORM_URL: JSON.stringify(process.env.PLATFORM_URL || 'https://pilot.traefik.io')
        }
        : { // production:
          APP_ENV: JSON.stringify(process.env.APP_ENV),
          APP_API: JSON.stringify(process.env.APP_API || '/api'),
          PLATFORM_URL: JSON.stringify(process.env.PLATFORM_URL || 'https://pilot.traefik.io')
        },
      uglifyOptions: {
        compress: {
          drop_console: process.env.APP_ENV === 'production',
          drop_debugger: process.env.APP_ENV === 'production'
        }
      },
      scopeHoisting: true,
      // vueRouterMode: 'history',
      // vueCompiler: true,
      // gzip: true,
      // analyze: true,
      // extractCSS: false,
      extendWebpack (cfg) {
        cfg.module.rules.push({
          enforce: 'pre',
          test: /\.(js|vue)$/,
          loader: 'eslint-loader',
          exclude: /node_modules/,
          options: {
            formatter: require('eslint').CLIEngine.getFormatter('stylish')
          }
        })
      }
    },

    devServer: {
      // https: true,
      port: 8081,
      open: true, // opens browser window automatically
      proxy: {
        // proxy all API requests to real Traefik
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      }
    },

    // animations: 'all', // --- includes all animations
    animations: [],

    ssr: {
      pwa: false
    },

    pwa: {
      // workboxPluginMode: 'InjectManifest',
      // workboxOptions: {}, // only for NON InjectManifest
      workboxOptions: {
        skipWaiting: true,
        clientsClaim: true
      },
      manifest: {
        // name: 'Traefik',
        // short_name: 'Traefik',
        // description: 'Traefik UI',
        display: 'standalone',
        orientation: 'portrait',
        background_color: '#ffffff',
        theme_color: '#027be3',
        icons: [
          {
            'src': 'statics/icons/icon-128x128.png',
            'sizes': '128x128',
            'type': 'image/png'
          },
          {
            'src': 'statics/icons/icon-192x192.png',
            'sizes': '192x192',
            'type': 'image/png'
          },
          {
            'src': 'statics/icons/icon-256x256.png',
            'sizes': '256x256',
            'type': 'image/png'
          },
          {
            'src': 'statics/icons/icon-384x384.png',
            'sizes': '384x384',
            'type': 'image/png'
          },
          {
            'src': 'statics/icons/icon-512x512.png',
            'sizes': '512x512',
            'type': 'image/png'
          }
        ]
      }
    },

    cordova: {
      // id: 'us.containo.traefik',
      // noIosLegacyBuildFlag: true, // uncomment only if you know what you are doing
    },

    electron: {
      // bundler: 'builder', // or 'packager'

      extendWebpack (cfg) {
        // do something with Electron main process Webpack cfg
        // chainWebpack also available besides this extendWebpack
      },

      packager: {
        // https://github.com/electron-userland/electron-packager/blob/master/docs/api.md#options

        // OS X / Mac App Store
        // appBundleId: '',
        // appCategoryType: '',
        // osxSign: '',
        // protocol: 'myapp://path',

        // Windows only
        // win32metadata: { ... }
      },

      builder: {
        // https://www.electron.build/configuration/configuration

        // appId: 'traefik-ui'
      }
    }
  }
}
