<template>
  <el-row :gutter="24">
    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <span>CloudSaver 配置</span>
            </div>
          </div>
        </template>
        <el-form label-position="top">
          <el-form-item label="服务地址">
            <el-input v-model="searchConfig.cloudsaver.server" placeholder="http://localhost:8080" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="searchConfig.cloudsaver.username" placeholder="用户名" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="searchConfig.cloudsaver.password" type="password" show-password placeholder="密码" />
          </el-form-item>
          <el-form-item label="Token 状态">
            <el-tag :type="searchConfig.cloudsaver.token ? 'success' : 'info'">
              {{ searchConfig.cloudsaver.token ? '已获取' : '未获取' }}
            </el-tag>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <span>PanSou 配置</span>
            </div>
          </div>
        </template>
        <el-form label-position="top">
          <el-form-item label="服务地址">
            <el-input v-model="searchConfig.pansou.server" placeholder="https://so.252035.xyz" />
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>

  <div style="text-align: right; margin-top: 16px;">
    <el-button type="primary" @click="saveSearchConfig" :loading="searchConfigSaving">
      保存配置
    </el-button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getSearchConfig, updateSearchConfig } from '../../api/search'
import { ElMessage } from 'element-plus'

const searchConfig = ref({
  cloudsaver: { server: '', username: '', password: '', token: '' },
  pansou: { server: '' }
})
const searchConfigSaving = ref(false)

const loadSearchConfig = async () => {
  try {
    const data = await getSearchConfig()
    if (data) {
      searchConfig.value = data
    }
  } catch (e) {
    console.error('加载搜索配置失败:', e)
  }
}

const saveSearchConfig = async () => {
  searchConfigSaving.value = true
  try {
    await updateSearchConfig(searchConfig.value)
    ElMessage.success('搜索配置已保存')
  } catch (e) {
    console.error('保存搜索配置失败:', e)
  } finally {
    searchConfigSaving.value = false
  }
}

onMounted(() => {
  loadSearchConfig()
})
</script>

<style scoped>
.inner-settings-card {
  background: var(--surface-bg) !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
  margin-bottom: 16px;
  border-radius: 12px !important;
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  font-size: 16px;
  color: var(--text-primary);
}
</style>
