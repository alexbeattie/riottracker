import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import MapView from '../views/MapView.vue'
import NewRioterView from '../views/NewRioterView.vue'  // <-- Import the new view

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView
  },
  {
    path: '/map',
    name: 'map',
    component: MapView
  },
  {
    path: '/new',
    name: 'new',
    component: NewRioterView  // <-- New route for the form
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router