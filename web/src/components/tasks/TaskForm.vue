<template>
  <!-- 创建/编辑任务抽屉 -->
  <el-drawer v-model="dialogVisible" :title="localForm.id ? '编辑任务' : '创建新任务'" direction="rtl" size="560px" destroy-on-close>
    <el-form :model="localForm" label-position="top" ref="formRef">
      <el-form-item label="智能粘贴解析" v-if="!localForm.id">
        <el-input
          v-model="smartInput"
          type="textarea"
          :rows="3"
          placeholder="请在此粘贴包含分享链接和提取码的文字，系统将自动尝试解析并填充下方表单"
          @input="handleSmartInput"
        />
      </el-form-item>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="任务名称" required>
            <el-input v-model="localForm.name" placeholder="给任务起个名字" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="执行账号" required>
            <el-select v-model="localForm.account_id" placeholder="选择账号" style="width: 100%" @change="onAccountChange">
              <el-option-group
                v-for="group in groupedAccounts"
                :key="group.label"
                :label="group.label"
              >
                <el-option
                  v-for="acc in group.options"
                  :key="acc.id"
                  :value="acc.id"
                  :disabled="acc.status === 0"
                  :label="acc.nickname"
                >
                  <div class="account-option-item">
                    <div class="acc-info">
                      <el-icon class="acc-icon" :color="acc.platform === 'quark' ? 'var(--color-quark)' : 'var(--color-139)'">
                        <PhCloud weight="duotone" />
                      </el-icon>
                      <span class="acc-name">{{ acc.nickname }}</span>
                    </div>
                    <div class="acc-meta">
                      <span class="acc-cap" v-if="acc.capacity_total > 0">
                        剩余 {{ formatSize(acc.capacity_total - acc.capacity_used) }}
                      </span>
                      <el-tag v-if="acc.status === 0" size="small" type="danger">已失效</el-tag>
                    </div>
                  </div>
                </el-option>
              </el-option-group>
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="分享链接" required>
        <div class="share-url-row">
          <el-input v-model="localForm.share_url" placeholder="请输入 139 或 Quark 分享链接" @change="onUrlChange" />
          <el-button-group class="share-url-actions">
            <el-button
              :icon="PhFolderOpen"
              title="浏览分享内容并选择目录"
              :disabled="!localForm.share_url || !localForm.account_id"
              @click="openBrowseShareDialog"
            />
            <el-button
              :icon="PhArrowSquareOut"
              title="在新标签页中打开链接"
              :disabled="!localForm.share_url"
              @click="emit('open-external', localForm.share_url, localForm.extract_code)"
            />
            <el-button
              type="primary"
              title="搜索替换资源"
              @click="openSearchReplace"
            >
              搜索替换
            </el-button>
          </el-button-group>
        </div>
      </el-form-item>

      <div v-if="isSubDirMode" class="subdir-hint">
        <el-tag type="warning" closable @close="emit('reset-share-root')">
          <el-icon style="margin-right: 4px; vertical-align: middle;"><PhFolderOpen /></el-icon>
          当前目录：{{ selectedDirName || '子目录' }}
        </el-tag>
      </div>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="提取码">
            <el-input v-model="localForm.extract_code" placeholder="如果有提取码请填写" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="起始转存点 (可选)">
        <div class="path-input-group">
          <el-input
            v-model="selectedStartFileName"
            placeholder="从该文件开始向前转存 (为空则转存全量)"
            readonly
            class="save-path-input"
          >
            <template #append>
              <el-button @click="openStartFileDialog" :loading="parsingShare">选择文件</el-button>
            </template>
          </el-input>
        </div>
        <div class="start-id-tip" v-if="localForm.start_file_id">
          <el-icon><PhInfo /></el-icon>
          <span>当前已锁定 ID: <code class="id-code" :title="localForm.start_file_id">{{ localForm.start_file_id }}</code></span>
        </div>
      </el-form-item>

      <el-form-item label="保存路径" required>
        <div class="path-input-group">
          <el-input
            v-model="localForm.save_path"
            placeholder="可手动输入或点击右侧选择"
            class="save-path-input"
          >
            <template #prepend v-if="selectedAccountPlatform">
              [{{ selectedAccountPlatform }}]
            </template>
            <template #append>
              <el-button @click="openFolderDialog">选择目录</el-button>
            </template>
          </el-input>
        </div>
      </el-form-item>

      <el-divider>调度规则与预览</el-divider>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="调度模式" required>
            <div class="schedule-mode-selector">
              <el-radio-group v-model="localForm.schedule_mode" @change="handleScheduleModeChange">
                <el-radio label="global">跟随全局</el-radio>
                <el-radio label="custom">自定义频率</el-radio>
                <el-radio label="off">手动执行</el-radio>
              </el-radio-group>
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20" v-if="localForm.schedule_mode === 'custom'">
        <el-col :span="24">
          <el-form-item label="自定义频率 (Cron)">
            <div style="display: flex; align-items: center; gap: 15px; width: 100%;">
              <el-select v-model="cronPreset" placeholder="选择预设频率" style="width: 180px" @change="handleCronPreset">
                <el-option label="每小时" value="0 0 * * * *" />
                <el-option label="每 6 小时" value="0 0 */6 * * *" />
                <el-option label="每天凌晨 2 点" value="0 0 2 * * *" />
                <el-option label="每周一凌晨 2 点" value="0 0 2 * * 1" />
                <el-option label="使用自定义表达式" value="custom" />
              </el-select>
              <el-input v-if="cronPreset === 'custom'" v-model="localForm.cron" placeholder="秒 分 时 日 月 周" style="flex: 1">
                <template #suffix>
                  <el-tooltip placement="top">
                    <template #content>
                      格式：秒 分 时 日 月 周 (6位)<br/>
                      例如: */5 * * * * * (每5秒执行一次)
                    </template>
                    <el-icon style="cursor: help"><PhInfo /></el-icon>
                  </el-tooltip>
                </template>
              </el-input>
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="重命名正则 (Pattern)">
            <el-input v-model="localForm.pattern" placeholder="匹配文件名的正则表达式" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="替换规则 (Replacement)">
            <el-input v-model="localForm.replacement" placeholder="支持 {TASKNAME}, {YEAR} 等变量" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="最大重试次数">
            <el-input-number v-model="localForm.max_retries" :min="1" :max="10" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="忽略后缀去重">
            <el-switch v-model="localForm.ignore_extension" active-text="开启" inactive-text="关闭" />
            <div class="form-tip">开启后 01.mp4 和 01.mkv 视为同一文件</div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="文件过滤规则">
            <el-input v-model="localForm.filter" placeholder="正则表达式，匹配的文件将被跳过" clearable />
            <div class="form-tip">留空表示不过滤；匹配到的文件将被跳过不转存</div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="起始日期过滤">
            <el-date-picker v-model="localForm.start_date" type="date" placeholder="仅转存此日期之后的文件" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ssZ" clearable />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="运行星期">
        <el-checkbox-group v-model="runDaysSelected">
          <el-checkbox v-for="d in dayOptions" :key="d.value" :label="d.value">{{ d.label }}</el-checkbox>
        </el-checkbox-group>
        <div class="form-tip">不选表示每天运行；选中表示仅在指定星期运行</div>
      </el-form-item>

      <div class="preview-action">
        <el-button type="success" :icon="PhArrowsClockwise" @click="emit('preview')" :loading="previewLoading">全量重命名预览</el-button>
      </div>
    </el-form>

    <template #footer>
      <div style="flex: auto">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="emit('submit')">确认并保存</el-button>
      </div>
    </template>
  </el-drawer>

  <!-- 目录选择独立弹窗 -->
  <el-dialog
    v-model="folderDialogVisible"
    title="选择保存目录"
    width="600px"
    class="folder-dialog"
    append-to-body
    destroy-on-close
  >
    <div class="folder-tree-container" v-loading="loadingFolders">
      <el-tree
        ref="folderTreeRef"
        lazy
        :load="loadFolders"
        :props="{ label: 'label', isLeaf: 'isLeaf' }"
        node-key="path"
        :default-expanded-keys="['/']"
        highlight-current
        @current-change="handleTreeCurrentChange"
        :expand-on-click-node="false"
        :empty-text="loadingFolders ? '加载中...' : '暂无目录'"
      >
        <template #default="{ node }">
          <span class="custom-tree-node">
            <el-icon><PhFolder /></el-icon>
            <span>{{ node.label }}</span>
          </span>
        </template>
      </el-tree>
    </div>
    <template #footer>
      <div class="folder-dialog-footer">
        <div class="create-folder-action">
          <el-input v-model="newFolderName" placeholder="新文件夹名称" size="default">
            <template #append>
              <el-button @click="handleInlineCreateFolder" :loading="creatingFolder">新建</el-button>
            </template>
          </el-input>
        </div>
        <div class="dialog-actions">
          <el-button @click="folderDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmFolderSelection">确认路径</el-button>
        </div>
      </div>
    </template>
  </el-dialog>

  <!-- 选择起始文件 / 浏览分享内容弹窗 -->
  <el-dialog
    v-model="startFileDialogVisible"
    :title="browseMode === 'selectShareUrl' ? '浏览分享内容' : '选择起始转存文件'"
    width="900px"
    append-to-body
    destroy-on-close
  >
    <div class="share-files-dialog-content">
      <!-- 面包屑导航 -->
      <div class="breadcrumb-nav" style="margin-bottom: 12px;">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>
            <a href="#" @click.prevent="navigateToBreadcrumb(-1)" class="breadcrumb-link">根目录</a>
          </el-breadcrumb-item>
          <el-breadcrumb-item v-for="(crumb, index) in breadcrumbs" :key="crumb.id">
            <a href="#" @click.prevent="navigateToBreadcrumb(index)" class="breadcrumb-link">{{ crumb.name }}</a>
          </el-breadcrumb-item>
        </el-breadcrumb>
        <el-button
          v-if="isSubDirMode"
          type="warning"
          link
          size="small"
          @click="emit('reset-share-root'); startFileDialogVisible = false"
          style="margin-left: auto;"
        >
          重置为根目录
        </el-button>
      </div>

      <el-alert v-if="browseMode === 'selectShareUrl'" title="选择目录" type="info" :closable="false" show-icon style="margin-bottom: 15px">
        点击文件夹进入子目录，选择目标文件夹后点击"选择为分享链接"，将更新分享链接为该目录的地址。
      </el-alert>
      <el-alert v-else title="逻辑说明" type="info" :closable="false" show-icon style="margin-bottom: 15px">
        系统将从选中的文件开始，按更新时间向前转存所有更新的文件（含所选文件本身）。<b>此处已应用您的重命名规则并执行同名预检。</b> 点击文件夹可进入子目录。
      </el-alert>

      <el-table
        :data="shareFiles"
        max-height="500"
        size="default"
        border
        stripe
        v-loading="parsingShare"
        highlight-current-row
        :row-class-name="tableRowClassName"
        @current-change="handleStartFileTableChange"
        @row-dblclick="handleRowDblClick"
      >
        <!-- startFile 模式且在初始目录时显示 radio 列 -->
        <el-table-column v-if="browseMode === 'startFile' && isInitialDir" width="40" align="center">
          <template #default="{ row }">
            <el-radio v-if="!row.is_folder" v-model="tempStartFileId" :label="row.id" class="naked-radio"><span></span></el-radio>
          </template>
        </el-table-column>

        <el-table-column label="原始文件名" show-overflow-tooltip min-width="180">
          <template #default="{ row }">
            <div class="name-main" :class="{ 'folder-clickable': row.is_folder }" @click="row.is_folder && enterFolder(row)" @dblclick="!row.is_folder && handleRowDblClick(row)">
              <el-icon size="16">
                <PhFolder v-if="row.is_folder" color="var(--color-warning)" />
                <PhFile v-else color="var(--text-muted)" />
              </el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="预览文件名 (入库名)" show-overflow-tooltip min-width="220">
          <template #default="{ row }">
            <span :style="{
              fontWeight: row.is_folder ? '600' : 'normal',
              color: (row.new_name && row.new_name !== row.name) ? 'var(--el-color-primary)' : 'inherit'
            }">
              {{ row.new_name || row.name }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.is_folder ? 'warning' : 'info'">
              {{ row.is_folder ? '目录' : '文件' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_existed" size="small" type="success">已在网盘</el-tag>
            <el-tag v-else size="small" type="info">待转存</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="updated_at" label="分享更新时间" width="160" sortable />

        <!-- selectShareUrl 模式显示进入按钮 -->
        <el-table-column v-if="browseMode === 'selectShareUrl'" label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button v-if="row.is_folder" type="primary" link size="small" @click="enterFolder(row)">
              进入
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <template #footer>
      <el-button @click="startFileDialogVisible = false">取消</el-button>
      <template v-if="browseMode === 'selectShareUrl'">
        <el-button type="primary" @click="confirmSelectShareUrl">
          选择当前目录（{{ currentDirName }}）
        </el-button>
      </template>
      <template v-else>
        <el-button @click="clearStartFile">清除选择</el-button>
        <el-button type="primary" @click="confirmStartFileSelection" :disabled="!tempStartFileId">确认选择</el-button>
      </template>
    </template>
  </el-dialog>

  <!-- 预览结果对话框 -->
  <el-dialog v-model="previewVisible" title="重命名预览" width="800px">
    <el-table :data="previewData" height="400px" border stripe>
      <el-table-column prop="original_name" label="原始文件名" min-width="200" show-overflow-tooltip />
      <el-table-column label="重命名后效果" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.is_filtered" class="filtered-text">(已过滤)</span>
          <span v-else-if="row.matched" class="matched-text">{{ row.new_name }}</span>
          <span v-else class="unmatched-text">{{ row.new_name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="row.is_filtered ? 'info' : (row.matched ? 'success' : 'warning')">
            {{ row.is_filtered ? '跳过' : (row.matched ? '匹配' : '未匹配') }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
    <div class="preview-tips">
      <p>* 列表基于当前分享链接的真实文件解析。</p>
      <p>* 过滤规则：如果设置了起始文件，则该文件之后（更旧）的文件将不执行转存。</p>
    </div>
  </el-dialog>

  <!-- 搜索替换弹窗 -->
  <el-dialog
    v-model="searchReplaceVisible"
    title="搜索替换资源"
    width="600px"
  >
    <div class="search-replace-bar">
      <el-input
        v-model="searchReplaceQuery"
        placeholder="输入搜索关键词"
        @keyup.enter="doSearchReplace"
      >
        <template #append>
          <el-button @click="doSearchReplace">搜索</el-button>
        </template>
      </el-input>
    </div>
    <div v-loading="searchReplaceLoading" class="search-replace-results">
      <div v-for="item in searchReplaceResults" :key="item.url" class="search-result-item">
        <div class="result-info">
          <span class="result-title">{{ item.title }}</span>
          <span class="result-source">{{ item.source }} - {{ item.platform }}</span>
        </div>
        <div class="result-actions">
          <el-button size="small" @click="emit('view-share-content', item)">查看内容</el-button>
          <el-button size="small" type="primary" @click="handleReplaceFromSearch(item)">替换</el-button>
        </div>
      </div>
      <el-empty v-if="!searchReplaceLoading && searchReplaceResults.length === 0" description="暂无搜索结果" />
    </div>
  </el-dialog>

  <!-- 分享内容弹窗 -->
  <ShareContentDialog
    v-model:visible="shareDialogVisible"
    :url="shareDialogUrl"
    :extract-code="shareDialogExtractCode"
    :title="shareDialogTitle"
    :show-replace="true"
    @replace-link="handleReplaceFromDialog"
  />
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  PhFolder, PhFile, PhInfo, PhCloud, PhArrowSquareOut,
  PhFolderOpen, PhArrowsClockwise
} from '@phosphor-icons/vue'
import { ElMessage } from 'element-plus'
import { searchResources } from '../../api/search'
import { parseShareLink } from '../../api/task'
import { getFolders, createFolder } from '../../api/account'
import ShareContentDialog from '../ShareContentDialog.vue'
import { formatSize } from '../../utils/format'
import { getDefaultFormData } from './utils'

const props = defineProps({
  /** 账号列表 */
  accounts: { type: Array, default: () => [] },
  /** 是否提交中 */
  submitting: { type: Boolean, default: false },
  /** 预览加载中 */
  previewLoading: { type: Boolean, default: false }
})

const emit = defineEmits([
  'submit',
  'open-external',
  'reset-share-root',
  'preview',
  'view-share-content'
])

// ---- 本地表单数据（由 TaskForm 自己管理）----
const localForm = ref(getDefaultFormData())
const selectedDirName = ref('')
const selectedStartFileName = ref('')
const pathIdMap = ref({ '/': '' })
const dialogVisible = ref(false)

// ---- 运行星期 ----
const dayOptions = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 7, label: '周日' },
]
const runDaysSelected = computed({
  get: () => {
    try { return JSON.parse(localForm.value.run_days || '[]') } catch { return [] }
  },
  set: (val) => { localForm.value.run_days = JSON.stringify(val) }
})

// ---- 智能解析 ----
const smartInput = ref('')

const handleSmartInput = (val) => {
  if (!val) return

  // 1. 提取链接
  const urlMatch = val.match(/https?:\/\/[a-zA-Z0-9.\-_/]+/)
  if (urlMatch) {
    localForm.value.share_url = urlMatch[0]
    onUrlChange()
  }

  // 2. 提取密码/提取码
  let remainText = val
  if (urlMatch) {
    remainText = val.replace(urlMatch[0], '')
  }

  const pwdKeywordMatch = remainText.match(/(?:提取码|提取|密码|pw|码|pwd)[:：\s]*([a-zA-Z0-9]+)/i)
  if (pwdKeywordMatch) {
    localForm.value.extract_code = pwdKeywordMatch[1]
  } else {
    const purePwdMatch = remainText.match(/\b([a-zA-Z0-9]{4})\b/)
    if (purePwdMatch) {
      localForm.value.extract_code = purePwdMatch[1]
    }
  }
}

// ---- 账号分组 ----
const groupedAccounts = computed(() => {
  if (!props.accounts) return []
  const groups = [
    { label: '移动云盘', platform: '139', options: [] },
    { label: '夸克网盘', platform: 'quark', options: [] }
  ]

  props.accounts.forEach(acc => {
    if (!acc) return
    const group = groups.find(g => g.platform === acc.platform)
    if (group) {
      group.options.push(acc)
    }
  })

  return groups.filter(g => g.options.length > 0)
})

const selectedAccountPlatform = computed(() => {
  if (!localForm.value.account_id || !props.accounts) return ''
  const account = props.accounts.find(acc => acc.id === localForm.value.account_id)
  if (account) {
    return account.platform === '139' ? '移动云盘' : 'Quark'
  }
  return ''
})

// 子目录模式判断
const isSubDirMode = computed(() => {
  const account = props.accounts.find(acc => acc.id === localForm.value.account_id)
  if (!account) return false

  if (account.platform === 'quark') {
    const match = localForm.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    return match && match[2] && match[2] !== '0'
  } else {
    return !!localForm.value.share_parent_id
  }
})

// ---- Cron 预设 ----
const cronPreset = ref('')

const handleScheduleModeChange = (val) => {
  if (val === 'custom' && !localForm.value.cron) {
    cronPreset.value = '0 0 * * * *'
    localForm.value.cron = '0 0 * * * *'
  }
}

const handleCronPreset = (val) => {
  if (val !== 'custom') {
    localForm.value.cron = val
  }
}

// ---- 链接变更处理 ----
const onUrlChange = () => {
  localForm.value.start_file_id = ''
  localForm.value.start_file_name = ''
  selectedStartFileName.value = ''
  localForm.value.share_parent_id = ''
  selectedDirName.value = ''
}

// ---- 切换账号处理 ----
const onAccountChange = () => {
  localForm.value.save_path = '/'
  localForm.value.start_file_id = ''
  localForm.value.start_file_name = ''
  selectedStartFileName.value = ''
  pathIdMap.value = { '/': '' }
}

// ---- 预览 ----
const previewVisible = ref(false)
const previewData = ref([])

// ---- 搜索替换 ----
const searchReplaceVisible = ref(false)
const searchReplaceQuery = ref('')
const searchReplaceResults = ref([])
const searchReplaceLoading = ref(false)

// ---- 分享内容弹窗 ----
const shareDialogVisible = ref(false)
const shareDialogUrl = ref('')
const shareDialogExtractCode = ref('')
const shareDialogTitle = ref('')

// ---- 目录选择弹窗 ----
const folderDialogVisible = ref(false)
const loadingFolders = ref(false)
const folderTreeRef = ref(null)
const selectedTreePath = ref('')
const selectedTreeId = ref('')
const newFolderName = ref('')
const creatingFolder = ref(false)

// ---- 分享文件弹窗 ----
const startFileDialogVisible = ref(false)
const shareFiles = ref([])
const parsingShare = ref(false)
const tempStartFileId = ref('')
const breadcrumbs = ref([])
const currentParentId = ref('')
const browseMode = ref('startFile')
const isInitialDir = ref(true)

// 面包屑目录名称
const currentDirName = computed(() => {
  if (breadcrumbs.value.length === 0) {
    return '根目录'
  }
  return breadcrumbs.value[breadcrumbs.value.length - 1].name
})

// 表格行样式
const tableRowClassName = ({ row }) => {
  if (row.is_existed) {
    return 'existed-row'
  }
  return ''
}

// ---- 目录树加载 ----
const loadFolders = async (node, resolve) => {
  if (!localForm.value.account_id) {
    return resolve([])
  }

  if (node.level === 0) {
    pathIdMap.value['/'] = ''
    return resolve([{ label: '根目录', path: '/', id: '', isLeaf: false }])
  }

  const parentID = node.data.id
  const parentPath = node.data.path

  if (node.level === 1) {
    loadingFolders.value = true
  }

  try {
    const folders = await getFolders(localForm.value.account_id, parentID, parentPath)

    setTimeout(() => {
      const newMappings = {}
      folders.forEach(f => {
        newMappings[f.path] = f.id
      })
      Object.assign(pathIdMap.value, newMappings)
    }, 0)

    resolve(folders)
  } catch (err) {
    console.error('加载目录失败:', err)
    resolve([])
  } finally {
    if (node.level === 1) {
      loadingFolders.value = false
    }
  }
}

const handleTreeCurrentChange = (data) => {
  if (data) {
    selectedTreePath.value = data.path
    selectedTreeId.value = data.id
  }
}

const confirmFolderSelection = () => {
  if (selectedTreePath.value) {
    localForm.value.save_path = selectedTreePath.value
  }
  folderDialogVisible.value = false
}

// ---- 分享文件加载 ----
const loadShareFiles = async (parentId) => {
  parsingShare.value = true
  shareFiles.value = []
  currentParentId.value = parentId

  try {
    const data = await parseShareLink({
      account_id: localForm.value.account_id,
      share_url: localForm.value.share_url,
      extract_code: localForm.value.extract_code,
      parent_id: parentId,
      save_path: localForm.value.save_path,
      pattern: localForm.value.pattern,
      replacement: localForm.value.replacement,
      name: localForm.value.name
    })
    shareFiles.value = data

    if (localForm.value.start_file_id) {
      const selected = data.find(f => f.id === localForm.value.start_file_id)
      if (selected) {
        selectedStartFileName.value = selected.name
      }
    }
  } catch (err) {
    console.error('解析链接失败:', err)
    startFileDialogVisible.value = false
  } finally {
    parsingShare.value = false
  }
}

const enterFolder = async (folder) => {
  breadcrumbs.value.push({ id: folder.id, name: folder.name })
  isInitialDir.value = false
  await loadShareFiles(folder.id)
}

const getInitialDirId = () => {
  const account = props.accounts.find(acc => acc.id === localForm.value.account_id)

  if (account?.platform === 'quark') {
    const match = localForm.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    if (match && match[2] && match[2] !== '0') {
      return match[2]
    }
  } else if (account?.platform === '139') {
    return localForm.value.share_parent_id || ''
  }

  return ''
}

const navigateToBreadcrumb = async (index) => {
  if (index === -1) {
    breadcrumbs.value = []
    const rootId = getInitialDirId()
    currentParentId.value = rootId
    isInitialDir.value = true
    await loadShareFiles(rootId)
  } else {
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    const navigatedId = breadcrumbs.value[index].id
    currentParentId.value = navigatedId
    const initialDirId = getInitialDirId()
    isInitialDir.value = initialDirId
      ? navigatedId === initialDirId
      : breadcrumbs.value.length === 0
    await loadShareFiles(navigatedId)
  }
}

const handleStartFileTableChange = (row) => {
  if (row) {
    tempStartFileId.value = row.id
  }
}

const handleRowDblClick = (row) => {
  if (browseMode.value === 'startFile' && !row.is_folder && isInitialDir.value) {
    tempStartFileId.value = row.id
    confirmStartFileSelection()
  }
}

const confirmStartFileSelection = () => {
  if (tempStartFileId.value) {
    localForm.value.start_file_id = tempStartFileId.value
    const selected = shareFiles.value.find(f => f.id === tempStartFileId.value)
    if (selected) {
      localForm.value.start_file_name = selected.name
      selectedStartFileName.value = selected.name
    }
  }
  startFileDialogVisible.value = false
}

// 确认选择目录作为新的分享链接
const confirmSelectShareUrl = () => {
  const account = props.accounts.find(acc => acc.id === localForm.value.account_id)
  if (!account) return

  const originalUrl = localForm.value.share_url
  let newUrl = originalUrl
  const currentDirId = currentParentId.value || ''

  if (account.platform === 'quark') {
    const match = originalUrl.match(/\/s\/(\w+)/)
    if (match) {
      const pwdID = match[1]
      const pdirFID = currentDirId || '0'
      newUrl = `https://pan.quark.cn/s/${pwdID}#/list/share/${pdirFID}`
    }
    localForm.value.share_parent_id = ''
  } else if (account.platform === '139') {
    localForm.value.share_parent_id = currentDirId || ''
    newUrl = originalUrl
  }

  localForm.value.share_url = newUrl
  localForm.value.start_file_id = ''
  localForm.value.start_file_name = ''
  selectedStartFileName.value = ''
  selectedDirName.value = currentDirName.value

  ElMessage.success(`已选择目录：${currentDirName.value}`)
  startFileDialogVisible.value = false
}

const clearStartFile = () => {
  localForm.value.start_file_id = ''
  localForm.value.start_file_name = ''
  selectedStartFileName.value = ''
  tempStartFileId.value = ''
  startFileDialogVisible.value = false
}

// ---- 打开目录选择弹窗 ----
const openFolderDialog = () => {
  if (!localForm.value.account_id) {
    return ElMessage.warning('请先选择执行账号')
  }
  loadingFolders.value = true
  selectedTreePath.value = localForm.value.save_path || '/'
  selectedTreeId.value = pathIdMap.value[localForm.value.save_path] || (localForm.value.save_path === '/' ? '' : '')
  newFolderName.value = ''
  folderDialogVisible.value = true
}

// ---- 打开起始文件弹窗 ----
const openStartFileDialog = async () => {
  if (!localForm.value.share_url || !localForm.value.account_id) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }
  browseMode.value = 'startFile'
  startFileDialogVisible.value = true
  parsingShare.value = true
  tempStartFileId.value = localForm.value.start_file_id
  shareFiles.value = []
  breadcrumbs.value = []
  isInitialDir.value = true
  const initialParentId = localForm.value.share_parent_id || ''
  currentParentId.value = initialParentId
  await loadShareFiles(initialParentId)
}

// ---- 打开浏览分享内容弹窗 ----
const openBrowseShareDialog = async () => {
  if (!localForm.value.share_url || !localForm.value.account_id) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }
  browseMode.value = 'selectShareUrl'
  startFileDialogVisible.value = true
  parsingShare.value = true
  tempStartFileId.value = ''
  shareFiles.value = []
  breadcrumbs.value = []

  const account = props.accounts.find(acc => acc.id === localForm.value.account_id)
  let initialParentId = ''

  if (account?.platform === 'quark') {
    const match = localForm.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    if (match && match[2] && match[2] !== '0') {
      initialParentId = match[2]
    }
  } else if (account?.platform === '139') {
    initialParentId = localForm.value.share_parent_id || ''
  }

  currentParentId.value = initialParentId
  await loadShareFiles(initialParentId)
}

// ---- 内嵌新建文件夹处理 ----
const handleInlineCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    return ElMessage.warning('请输入文件夹名称')
  }

  const currentPath = selectedTreePath.value || '/'
  const currentID = selectedTreeId.value || ''

  if (folderTreeRef.value) {
    const currentNode = folderTreeRef.value.getNode(currentPath)
    if (currentNode && currentNode.childNodes) {
      const isDuplicate = currentNode.childNodes.some(
        child => child.data && child.data.label === newFolderName.value.trim()
      )
      if (isDuplicate) {
        return ElMessage.warning('该目录下已存在同名文件夹')
      }
    }
  }

  creatingFolder.value = true
  try {
    const res = await createFolder(localForm.value.account_id, currentID, currentPath, newFolderName.value.trim())
    ElMessage.success('文件夹创建成功')
    pathIdMap.value[res.path] = res.id

    if (folderTreeRef.value) {
      const currentNode = folderTreeRef.value.getNode(currentPath)
      if (currentNode) {
        folderTreeRef.value.append(res, currentNode)
        currentNode.expanded = true
      }
      selectedTreePath.value = res.path
      selectedTreeId.value = res.id
      folderTreeRef.value.setCurrentKey(res.path)
    }
    newFolderName.value = ''
  } catch (err) {
    console.error('创建文件夹失败:', err)
  } finally {
    creatingFolder.value = false
  }
}

// ---- 搜索替换 ----
const openSearchReplace = () => {
  searchReplaceQuery.value = localForm.value.name
  searchReplaceResults.value = []
  searchReplaceVisible.value = true
  if (searchReplaceQuery.value) {
    doSearchReplace()
  }
}

const doSearchReplace = async () => {
  if (!searchReplaceQuery.value.trim()) return
  searchReplaceLoading.value = true
  try {
    const data = await searchResources({ q: searchReplaceQuery.value })
    searchReplaceResults.value = data.items || []
  } catch (e) {
    ElMessage.error('搜索失败')
  } finally {
    searchReplaceLoading.value = false
  }
}

const handleReplaceFromSearch = (item) => {
  localForm.value.share_url = item.url
  searchReplaceVisible.value = false
  ElMessage.success('链接已替换，请保存任务')
}

const handleReplaceFromDialog = (data) => {
  localForm.value.share_url = data.url
  localForm.value.extract_code = data.extractCode || ''
  shareDialogVisible.value = false
  ElMessage.success('链接已替换，请保存任务')
}

// ---- 暴露给主控的方法 ----
defineExpose({
  /** 获取当前表单数据 */
  getFormData: () => localForm.value,
  /** 打开抽屉（新建模式） */
  open(formData) {
    localForm.value = { ...formData }
    selectedStartFileName.value = ''
    selectedDirName.value = ''
    pathIdMap.value = { '/': '' }
    smartInput.value = ''
    cronPreset.value = ''
    dialogVisible.value = true
  },
  /** 打开抽屉（编辑模式） */
  openEdit(formData, startFileName, dirName) {
    localForm.value = { ...formData }
    selectedStartFileName.value = startFileName || ''
    selectedDirName.value = dirName || ''
    smartInput.value = ''

    // 初始化定时配置状态
    if (formData.schedule_mode === 'custom' && formData.cron) {
      const presets = ['0 0 * * * *', '0 0 */6 * * *', '0 0 2 * * *', '0 0 2 * * 1']
      cronPreset.value = presets.includes(formData.cron) ? formData.cron : 'custom'
    } else {
      cronPreset.value = ''
    }

    dialogVisible.value = true
  },
  /** 设置 Cron 预设 */
  setCronPreset(val) { cronPreset.value = val },
  /** 设置预览数据 */
  setPreviewData(data) {
    previewData.value = data
    previewVisible.value = true
  },
  /** 打开分享内容弹窗 */
  openShareContentDialog(item) {
    shareDialogUrl.value = item.url
    shareDialogExtractCode.value = ''
    shareDialogTitle.value = item.title
    shareDialogVisible.value = true
  },
  /** 关闭抽屉 */
  close() {
    dialogVisible.value = false
  },
  /** 抽屉是否打开 */
  isOpen() {
    return dialogVisible.value
  }
})
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.schedule-mode-selector {
  margin-bottom: 10px;
}

.path-input-group {
  display: flex;
  gap: 12px;
  align-items: center;
}

.start-id-tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--neutral-500);
  display: flex;
  align-items: center;
  gap: 4px;
}

.id-code {
  background-color: var(--neutral-100);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
  color: var(--brand-600);
  max-width: 300px;
  display: inline-block;
  vertical-align: bottom;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-files-dialog-content {
  padding: 10px 0;
}

.save-path-input {
  width: 100%;
}

.save-path-input :deep(.el-input-group__prepend) {
  background-color: var(--neutral-100);
  color: var(--neutral-600);
  font-weight: 700;
}

.folder-tree-container {
  height: 400px;
  overflow-y: auto;
  border: 1px solid var(--neutral-200);
  border-radius: 8px;
  padding: 8px;
}

.folder-dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.create-folder-action {
  flex: 1;
  margin-right: 20px;
  max-width: 300px;
}

.dialog-actions {
  flex-shrink: 0;
}

.preview-action {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

.preview-tips {
  margin-top: 16px;
  color: var(--neutral-500);
  font-size: 13px;
}

.preview-tips p {
  margin: 4px 0;
}

.filtered-text {
  color: var(--neutral-400);
  font-style: italic;
}

.matched-text {
  color: var(--color-success);
  font-weight: 500;
}

.unmatched-text {
  color: var(--color-warning);
}

.existed-row {
  background-color: var(--neutral-100) !important;
  color: var(--neutral-400);
}

.existed-row span {
  opacity: 0.8;
}

:deep(.el-tree-node__children) {
  transition: none !important;
}
:deep(.el-collapse-transition-leave-active),
:deep(.el-collapse-transition-enter-active) {
  transition: none !important;
  display: none !important;
}

.naked-radio :deep(.el-radio__label) {
  display: none !important;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-main {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.account-option-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  gap: 12px;
}

.acc-info {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.acc-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.acc-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.acc-cap {
  font-size: 12px;
  color: var(--neutral-400);
}

.breadcrumb-nav {
  padding: 8px 0;
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  color: var(--brand-600);
  cursor: pointer;
  text-decoration: none;
  font-weight: 500;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.folder-clickable {
  cursor: pointer;
}

.folder-clickable:hover {
  color: var(--brand-600);
}

.share-url-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.share-url-row .el-input {
  flex: 1;
}

.share-url-actions .el-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.subdir-hint {
  margin-top: 6px;
}

.subdir-hint .el-tag {
  cursor: default;
}

.search-replace-bar {
  margin-bottom: 16px;
}

.search-replace-results {
  min-height: 200px;
  max-height: 400px;
  overflow-y: auto;
}

.search-result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.result-info {
  flex: 1;
  overflow: hidden;
}

.result-title {
  display: block;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-source {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.result-actions {
  display: flex;
  gap: 8px;
  margin-left: 16px;
}
</style>
