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
