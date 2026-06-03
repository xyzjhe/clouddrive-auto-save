<template>
  <el-config-provider>
    <router-view />
    <PWAInstall />
  </el-config-provider>
</template>

<script setup>
import { ElConfigProvider } from 'element-plus'
import PWAInstall from './components/PWAInstall.vue'
</script>

<style>
#app {
  margin: 0;
  padding: 0;
}
</style>
