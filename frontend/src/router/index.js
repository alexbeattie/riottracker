import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import MapView from '../views/MapView.vue'
import NewRioterView from '../views/NewRioterView.vue'  // <-- Import the new view
import EditRioterView from '@/views/EditRioterView.vue'

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
  },
  {
    path: '/rioter/:id/edit',
    name: 'editRioter',
    component: EditRioterView
  },
  { path: "/edit", component: EditRioterView }, // ✅ Now it's valid

  // {
  //   path: '/update',
  //   name: 'update',
  //   component: UpdateRioterView  // <-- New route for the form
  // }

]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router