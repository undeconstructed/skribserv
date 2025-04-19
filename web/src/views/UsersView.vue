<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'

const app = useApp()

const datumoj = ref({ ento: null, eraro: '' })

app.getEntity('/uzantoj')
  .then(e => {
    datumoj.value.ento = e
  })
  .catch(err => {
    datumoj.value.eraro = err
  })
</script>

<template>
  <h1>Uzantoj</h1>

  <template v-if="datumoj.ento">
    <p><RouterLink :to="{ name: 'newuser' }">Krei uzanton</RouterLink></p>
    <p>Ĉiuj uzantoj:</p>
    <ul>
      <li v-for="k of datumoj.ento">
        <RouterLink :to="{ name: 'user', params: { id: k.id } }">{{ k.nomo }}</RouterLink>
      </li>
    </ul>
  </template>
  <p v-else-if="datumoj.eraro">{{ datumoj.eraro }}</p>
  <p v-else>Ŝarganta...</p>

</template>
