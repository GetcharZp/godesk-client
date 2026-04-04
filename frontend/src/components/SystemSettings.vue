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
  font-size: 24px;
  color: #00d4ff;
  font-weight: 500;
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
  letter-spacing: 0.5px;
}

.settings-form {
  background: linear-gradient(135deg, #151b3d 0%, #1a2040 100%);
  padding: 32px;
  border-radius: 12px;
  border: 1px solid #2d3561;
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.3);
  margin-bottom: 24px;
}

.form-item {
  margin-bottom: 28px;
}

.form-label {
  display: block;
  font-size: 14px;
  color: #e0e7ff;
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
  background: #0a0e27;
  border: 1px solid #2d3561;
  border-radius: 8px;
  font-size: 14px;
  color: #e0e7ff;
  transition: all 0.3s ease;
}

.form-input::placeholder {
  color: #64748b;
}

.form-input:focus {
  outline: none;
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}

.input-hint {
  font-size: 12px;
  color: #64748b;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}

.btn-save {
  padding: 12px 28px;
  background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
  color: #0a0e27;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
}

.btn-save:hover:not(:disabled) {
  box-shadow: 0 0 25px rgba(0, 212, 255, 0.5);
  transform: translateY(-2px);
}

.btn-save:disabled {
  background: linear-gradient(135deg, #3d4571 0%, #2d3561 100%);
  cursor: not-allowed;
  box-shadow: none;
}

.btn-reset {
  padding: 12px 28px;
  background: transparent;
  color: #94a3b8;
  border: 1px solid #2d3561;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-reset:hover {
  color: #00d4ff;
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}

.settings-info {
  background: rgba(16, 185, 129, 0.08);
  padding: 20px 24px;
  border-radius: 8px;
  border: 1px solid #10b981;
}

.settings-info h3 {
  font-size: 14px;
  color: #10b981;
  margin-bottom: 12px;
  text-shadow: 0 0 8px rgba(16, 185, 129, 0.3);
}

.settings-info ul {
  margin: 0;
  padding-left: 20px;
}

.settings-info li {
  font-size: 13px;
  color: #94a3b8;
  line-height: 1.8;
}
</style>
