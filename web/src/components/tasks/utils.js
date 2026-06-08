/**
 * 任务表单默认值工厂函数
 * 消除 Tasks.vue 中 3 处重复的表单初始值定义
 */
export function getDefaultFormData() {
  return {
    id: null,
    name: '',
    account_id: '',
    share_url: '',
    extract_code: '',
    save_path: '/',
    pattern: '',
    replacement: '',
    filter: '',
    start_file_id: '',
    start_file_name: '',
    share_parent_id: '',
    cron: '',
    schedule_mode: 'global',
    max_retries: 3,
    run_days: '',
    start_date: null,
    ignore_extension: false
  }
}
