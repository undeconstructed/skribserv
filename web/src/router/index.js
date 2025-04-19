import { createRouter, createWebHistory } from 'vue-router'
import { useApp } from '@/stores/app'

import HomeView from '@/views/HomeView.vue'
import UsersView from '@/views/UsersView.vue'
import UserView from '@/views/UserView.vue'
import CoursesView from '@/views/CoursesView.vue'
import CourseView from '@/views/CourseView.vue'
import NewCourseView from '@/views/NewCourseView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/uzantoj',
      name: 'users',
      component: UsersView,
    },
    {
      path: '/uzantoj/+nova',
      name: 'newuser',
      // component: NewCourseView,
    },
    {
      path: '/uzantoj/:id',
      name: 'user',
      component: UserView,
    },
    {
      path: '/kursoj',
      name: 'courses',
      component: CoursesView,
    },
    {
      path: '/kursoj/+nova',
      name: 'newcourse',
      component: NewCourseView,
    },
    {
      path: '/kursoj/:id',
      name: 'course',
      component: CourseView,
    },
    {
      path: '/pri',
      name: 'about',
      component: () => import('@/views/AboutView.vue'),
    },
  ],
})

// router.beforeEach((to) => {
//   const store = useTheStore()
//   if (to.name == 'login' && store.isLoggedIn) return '/'
//   if (to.meta.requiresAuth && !store.isLoggedIn) return '/login'
// })

export default router
