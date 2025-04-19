<script setup>
import { RouterView } from 'vue-router'
import Side from './components/Side.vue'
import LoginView from '@/views/LoginView.vue'
import { useApp } from '@/stores/app'

const app = useApp()
</script>

<template>
  <template v-if="app.isReady">
    <div v-if="app.session.isLoggedIn" id="normal">
      <header>
        <Side title="skribserv" />
      </header>

      <main>
        <RouterView />
      </main>
    </div>
    <div v-else id="login">
      <main>
        <LoginView />
      </main>
    </div>
  </template>
  <div v-else>
    <main>
      loading...
    </main>
  </div>
</template>

<style scoped>
#normal {
  height: 100%;
  width: 100%;

  display: grid;
  grid-template-columns: 15em 1fr;
  gap: 0px 0.5em;

}

#normal > header {
  background-color: lightgrey;
  padding: .5em;
}

#normal > main {
  overflow: auto;
  padding: .5em;
}

#login {
  height: 100%;
  width: 100%;
}
</style>
