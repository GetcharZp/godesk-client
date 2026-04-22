<template>
  <div v-if="visible" class="modal-overlay" @click.self="handleCancel">
    <div class="modal">
      <h3>{{ isEdit ? '编辑设备' : '添加设备' }}</h3>
      <div class="form">
        <div class="form-item">
          <label>设备码</label>
          <input type="number" v-model="formData.code" placeholder="请输入设备码" :disabled="isEdit">
        </div>
        <div class="form-item">
          <label>设备密码</label>
          <input type="password" v-model="formData.password" placeholder="请输入设备密码">
        </div>
        <div class="form-item">
          <label>备注</label>
          <input type="text" v-model="formData.remark" placeholder="请输入备注（可选）">
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-cancel" @click="handleCancel">取消</button>
        <button class="btn-confirm" @click="handleConfirm" :disabled="loading">
          {{ loading ? '保存中...' : '确认' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  isEdit: {
    type: Boolean,
    default: false
  },
  initialData: {
    type: Object,
    default: () => ({
      code: '',
      password: '',
      remark: ''
    })
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'confirm', 'cancel'])

const formData = ref({
  code: '',
  password: '',
  remark: ''
})

watch(() => props.visible, (newVal) => {
  if (newVal) {
    formData.value = {
      code: props.initialData.code || '',
      password: props.initialData.password || '',
      remark: props.initialData.remark || ''
    }
  }
})

const handleCancel = () => {
  emit('update:visible', false)
  emit('cancel')
}

const handleConfirm = () => {
  emit('confirm', {
    code: parseInt(formData.value.code) || 0,
    password: formData.value.password,
    remark: formData.value.remark
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-modal-mask);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: var(--bg-modal);
  padding: 28px;
  border-radius: 16px;
  border: 1px solid var(--border-primary);
  width: 420px;
  max-width: 90%;
  box-shadow: var(--shadow-card);
}

.modal h3 {
  margin-bottom: 24px;
  font-size: 20px;
  color: var(--text-accent);
  text-shadow: 0 0 15px var(--accent-primary-glow-strong);
  letter-spacing: 0.5px;
}

.form {
  margin-bottom: 28px;
}

.form-item {
  margin-bottom: 20px;
}

.form-item label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--text-secondary);
  letter-spacing: 0.3px;
}

.form-item input {
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-primary);
  box-sizing: border-box;
  transition: all 0.3s ease;
}

.form-item input::placeholder {
  color: var(--text-muted);
}

.form-item input:focus {
  outline: none;
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
}

.form-item input:disabled {
  background-color: var(--bg-tertiary);
  color: var(--text-muted);
  cursor: not-allowed;
  border-color: var(--border-secondary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 10px 20px;
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-cancel:hover {
  color: var(--text-accent);
  border-color: var(--border-active);
  box-shadow: 0 0 10px var(--accent-primary-glow);
}

.btn-confirm {
  padding: 10px 24px;
  background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-primary-dark) 100%);
  color: var(--text-on-accent);
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-button);
}

.btn-confirm:hover {
  box-shadow: 0 0 25px var(--accent-primary-glow-strong);
  transform: translateY(-2px);
}

.btn-confirm:disabled {
  background: linear-gradient(135deg, var(--scrollbar-thumb) 0%, var(--border-primary) 100%);
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}
</style>
