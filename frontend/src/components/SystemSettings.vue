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

      <!-- AccessToken 配置 -->
      <div class="form-item">
        <label class="form-label">AccessToken</label>
        <div class="form-input-wrapper">
          <input type="password" v-model="config.access_token" placeholder="请输入 AccessToken"
            class="form-input" />
          <span class="input-hint">用于后端请求验证，为空时不发送</span>
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
  service_address: '',
  access_token: ''
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
  max-width: 800px;
  margin: 0 auto;
}

.header {
  margin-bottom: 32px;
}

.page-title {
  font-size: 20px;
  color: var(--text-accent);
  font-weight: 500;
  text-shadow: 0 0 15px var(--accent-primary-glow-strong);
  letter-spacing: 0.5px;
}

.settings-form {
  background: var(--bg-card);
  padding: 32px;
  border-radius: 12px;
  border: 1px solid var(--border-primary);
  box-shadow: var(--shadow-glow);
  margin-bottom: 24px;
}

.form-item {
  margin-bottom: 28px;
}

.form-label {
  display: block;
  font-size: 14px;
  color: var(--text-primary);
  margin-bottom: 10px;
  font-weight: 500;
  letter-spacing: 0.3px;
}

.form-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-input {
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.form-input::placeholder {
  color: var(--text-muted);
}

.form-input:focus {
  outline: none;
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
}

.input-hint {
  font-size: 12px;
  color: var(--text-muted);
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}

.btn-save {
  padding: 12px 28px;
  background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-primary-dark) 100%);
  color: var(--text-on-accent);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-button);
}

.btn-save:hover:not(:disabled) {
  box-shadow: 0 0 25px var(--accent-primary-glow-strong);
  transform: translateY(-2px);
}

.btn-save:disabled {
  background: linear-gradient(135deg, var(--scrollbar-thumb) 0%, var(--border-primary) 100%);
  cursor: not-allowed;
  box-shadow: none;
}

.btn-reset {
  padding: 12px 28px;
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-reset:hover {
  color: var(--text-accent);
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
}

.settings-info {
  background: var(--success-bg);
  padding: 20px 24px;
  border-radius: 8px;
  border: 1px solid var(--success);
}

.settings-info h3 {
  font-size: 14px;
  color: var(--success);
  margin-bottom: 12px;
  text-shadow: 0 0 8px var(--success-glow);
}

.settings-info ul {
  margin: 0;
  padding-left: 20px;
}

.settings-info li {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.8;
}
</style>
