import Vue from 'vue'
import Router from 'vue-router'

// in development-env not use lazy-loading, because lazy-loading too many pages will cause webpack hot update too slow. so only in production use lazy-loading;
// detail: https://panjiachen.github.io/vue-element-admin-site/#/lazy-loading

Vue.use(Router)

/* Layout */
import Layout from '../views/layout/Layout'

/**
 * hidden: true                   if `hidden:true` will not show in the sidebar(default is false)
 * alwaysShow: true               if set true, will always show the root menu, whatever its child routes length
 *                                if not set alwaysShow, only more than one route under the children
 *                                it will becomes nested mode, otherwise not show the root menu
 * redirect: noredirect           if `redirect:noredirect` will no redirct in the breadcrumb
 * name:'router-name'             the name is used by <keep-alive> (must set!!!)
 * meta : {
    title: 'title'               the name show in submenu and breadcrumb (recommend set)
    icon: 'svg-name'             the icon show in the sidebar,
  }
 **/
export const constantRouterMap = [
  { path: '/login', component: () => import('@/views/login/index'), hidden: true },
  { path: '/404', component: () => import('@/views/404'), hidden: true },

  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    name: 'Dashboard',
    hidden: true,
    children: [{
      path: 'dashboard',
      component: () => import('@/views/dashboard/index')
    }]
  },

  {
    path: '/system',
    component: Layout,
    redirect: '/system/index',
    name: 'System',
    meta: { title: 'System', icon: 'example' },
    children: [
      {
        path: 'index',
        component: () => import('@/views/system/index'),
        meta: { title: '系统', icon: 'system' }
      }
    ]
  },

  {
    path: '',
    component: Layout,
    name: 'Cluster',
    meta: { title: '集群', icon: 'cluster', noCache: true },
    children: [
      {
        path: '/cluster/pepperbus',
        name: 'Cluster_Bus',
        component: () => import('@/views/cluster/index'),
        meta: { title: 'pepper_bus', icon: 'servers', noCache: true }
      },
      {
        path: '/cluster/peppercron',
        name: 'Cluster_Cron',
        component: () => import('@/views/cluster/index'),
        meta: { title: 'pepper_cron', icon: 'servers', noCache: true }
      }
    ]
  },
  {
    path: '/auth/list',
    component: Layout,
    redirect: '/auth/list',
    name: 'UserList',
    meta: { title: '用户列表', icon: 'example' },
    children: [
      {
        path: '/auth/list',
        component: () => import('@/views/user/index'),
        meta: { title: '用户列表', icon: 'queue' }
      }
    ],
    hidden: true
  },

  {
    path: '',
    component: Layout,
    name: 'Bus',
    meta: { title: 'Bus 总线', icon: 'queue' },
    children: [
      {
        path: '/bus/queue',
        name: 'Bus_Queue',
        component: () => import('@/views/queue/index'),
        meta: { title: '队列', icon: 'topic' }
      },
      {
        path: '/bus/queue-system',
        name: 'Bus_Queue',
        component: () => import('@/views/queue/index-system'),
        meta: { title: '队列-系统视图', icon: 'topic' }
      },
      {
        path: '/bus/storage',
        name: 'Bus_Storage',
        component: () => import('@/views/storage/index'),
        meta: { title: '存储', icon: 'storage' }
      }
    ]
  },

  {
    path: '',
    component: Layout,
    name: 'Cron',
    meta: { title: 'Cron 任务', icon: 'calendar' },
    children: [
      {
        path: '/cron/job',
        name: 'Cron_Job',
        component: () => import('@/views/job/index'),
        meta: { title: '任务', icon: 'tasks' }
      },
      {
        path: '/cron/node',
        name: 'Cron_Node',
        component: () => import('@/views/job/node'),

        meta: { title: '节点', icon: 'list' }

      },

      {
        path: '/cron/detail/:id',
        name: 'Detail',
        component: () => import('@/views/job/detail'),
        meta: { title: 'Detail', icon: 'storage' },
        hidden: true

        // hidden: true,

      }
    ]
  },

  {
    path: '/search',
    component: Layout,
    redirect: '/search/index',
    name: 'Search',
    meta: { title: 'Search', icon: 'eye' },
    children: [
      {
        path: 'index',
        component: () => import('@/views/topic/search'),
        meta: { title: '搜索', icon: 'eye' }
      }
    ]
  },

  {
    path: '/topic',
    component: Layout,
    redirect: '/topic/index',
    name: 'Topic',
    meta: { title: 'Topic', icon: 'example' },
    children: [
      {
        path: 'index',
        component: () => import('@/views/topic/index'),
        meta: { title: 'Topic', icon: 'table' }
      }
    ],
    hidden: true
  },

  // {
  //   path: 'external-link',
  //   component: Layout,
  //   children: [
  //     {
  //       path: 'https://panjiachen.github.io/vue-element-admin-site/#/',
  //       meta: { title: 'externalLink', icon: 'link' }
  //     }
  //   ]
  // },

  { path: '*', redirect: '/404', hidden: true }
]

export default new Router({
  // mode: 'history', //后端支持可开
  scrollBehavior: () => ({ y: 0 }),
  routes: constantRouterMap
})
