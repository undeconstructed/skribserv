<script setup>
import AdminHome from '../components/AdminHome.vue'
import UserHome from '../components/UserHome.vue'
import { useApp } from '@/stores/app'

const app = useApp()
</script>

<template>
  <h1>Hejmo</h1>
  <template v-if="app.session.admin">
    <AdminHome />
  </template>
  <template v-else>
    <UserHome />
  </template>
</template>
