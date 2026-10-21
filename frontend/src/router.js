import { createRouter, createWebHistory } from 'vue-router'
import ChartView from './views/ChartView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/chart/0' },
    { path: '/chart/:chartId', name: 'chart', component: ChartView }
  ]
})

export default router
