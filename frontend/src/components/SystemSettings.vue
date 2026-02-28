<template>
  <div class="system-settings">
    <div class="header">
      <h2 class="page-title">系统设置</h2>
    </div>

    <div class="settings-form">
      <!-- 服务地址配置 -->
      <div class="form-item">
        <label class="form-label">服务地址</label>
        <div class="form-input-wrapper">
          <input type="text" v-model="config.service_address" placeholder="请输入服务地址，例如：127.0.0.1:9620"
            class="form-input" />
          <span class="input-hint">默认地址：127.0.0.1:9620</span>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-save" @click="handleSave">
          {{ '保存设置' }}
        </button>
      </div>
    </div>

  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { getSysConfig, saveSysConfig, reconnect } from '../api/sys.js'

// 配置对象
const config = reactive({
  service_address: ''
})

onMounted(async () => {
  await loadSettings()
})

const loadSettings = async () => {
  try {
    const res = await getSysConfig()
    if (res && res.code === 200 && res.data) {
      Object.assign(config, res.data)
    }
  } catch (error) {

  }
}

const handleSave = async () => {
  if (!config.service_address.trim()) {
    message.error('请输入服务地址')
    return
  }

  try {
    const res = await saveSysConfig(config)
    if (res && res.code === 200) {
      await reconnect()
      message.success('保存成功')
    } else {
      message.error(res?.msg || '保存失败')
    }
  } catch (error) {
    message.error('保存失败：' + error.message)
  } finally {
  }
}

</script>

<style scoped>
.system-settings {
  padding: 20px;
  max-width: 600px;
}

.header {
  margin-bottom: 30px;
}

.page-title {
  font-size: 20px;
  color: #333;
  font-weight: 500;
}

.settings-form {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

.form-item {
  margin-bottom: 24px;
}

.form-label {
  display: block;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

.form-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-input {
  padding: 10px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.3s;
}

.form-input:focus {
  outline: none;
  border-color: #409EFF;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.input-hint {
  font-size: 12px;
  color: #999;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 30px;
}

.btn-save {
  padding: 10px 24px;
  background-color: #409EFF;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-save:hover:not(:disabled) {
  background-color: #66b1ff;
}

.btn-save:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.btn-reset {
  padding: 10px 24px;
  background-color: white;
  color: #666;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-reset:hover {
  color: #409EFF;
  border-color: #409EFF;
}

.settings-info {
  background: #f6ffed;
  padding: 16px 20px;
  border-radius: 8px;
  border: 1px solid #b7eb8f;
}

.settings-info h3 {
  font-size: 14px;
  color: #52c41a;
  margin-bottom: 10px;
}

.settings-info ul {
  margin: 0;
  padding-left: 20px;
}

.settings-info li {
  font-size: 13px;
  color: #666;
  line-height: 1.8;
}
</style>
