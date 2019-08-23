const LayoutDefault = () => import('layouts/Default.vue')

const routes = [
  {
    path: '/',
    component: LayoutDefault,
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('pages/dashboard/Index.vue'),
        meta: {
          title: 'Dashboard'
        }
      }
    ]
  }
  /* {
    path: '/http',
    redirect: '/http/routers',
    component: LayoutDefault,
    children: [
      {
        path: 'routers',
        name: 'httpRouters',
        components: {
          default: () => import('pages/http/Routers.vue'),
          NavBar: () => import('components/http/ToolBar.vue')
        },
        props: { default: true, NavBar: true },
        meta: {
          title: 'HTTP Routers'
        }
      },
      {
        path: 'services',
        name: 'httpServices',
        components: {
          default: () => import('pages/http/Services.vue'),
          NavBar: () => import('components/http/ToolBar.vue')
        },
        props: { default: true, NavBar: true },
        meta: {
          title: 'HTTP Services'
        }
      },
      {
        path: 'middlewares',
        name: 'httpMiddlewares',
        components: {
          default: () => import('pages/http/Middlewares.vue'),
          NavBar: () => import('components/http/ToolBar.vue')
        },
        props: { default: true, NavBar: true },
        meta: {
          title: 'HTTP Middlewares'
        }
      }
    ]
  },
  {
    path: '/tcp',
    redirect: '/tcp/routers',
    component: LayoutDefault,
    children: [
      {
        path: 'routers',
        name: 'tcpRouters',
        components: {
          default: () => import('pages/tcp/Routers.vue'),
          NavBar: () => import('components/tcp/ToolBar.vue')
        },
        props: { default: true, NavBar: true },
        meta: {
          title: 'TCP Routers'
        }
      },
      {
        path: 'services',
        name: 'tcpServices',
        components: {
          default: () => import('pages/tcp/Services.vue'),
          NavBar: () => import('components/tcp/ToolBar.vue')
        },
        props: { default: true, NavBar: true },
        meta: {
          title: 'TCP Services'
        }
      }
    ]
  } */
]

// Always leave this as last one
if (process.env.MODE !== 'ssr') {
  routes.push({
    path: '*',
    component: () => import('pages/_commons/Error404.vue'),
    meta: {
      title: '404'
    }
  })
}

export default routes
