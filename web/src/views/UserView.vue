<script setup>
import { ref } from 'vue'
import { useApp } from '@/stores/app'
import { useRoute } from 'vue-router'

const app = useApp()
const route = useRoute()

const datumoj = ref({ ento: null, eraro: '' })

app.getEntity('/uzantoj/' + route.params.id)
  .then(e => {
    datumoj.value.ento = e
  })
  .catch(err => {
    datumoj.value.eraro = err
  })
</script>

<template>

  <template v-if="datumoj.ento">
    <h1>Uzanto: {{ datumoj.ento.nomo }}</h1>
  </template>
  <template v-else-if="datumoj.eraro">
    <p>{{ datumoj.eraro }}</p>
  </template>
  <template v-else>
    <p>Ŝarganta...</p>
  </template>

</template>
