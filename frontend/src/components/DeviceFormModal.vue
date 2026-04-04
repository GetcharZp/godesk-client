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
  background-color: rgba(10, 14, 39, 0.85);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: linear-gradient(135deg, #151b3d 0%, #1a2040 100%);
  padding: 28px;
  border-radius: 16px;
  border: 1px solid #2d3561;
  width: 420px;
  max-width: 90%;
  box-shadow: 0 0 50px rgba(0, 0, 0, 0.5), 0 0 30px rgba(0, 212, 255, 0.1);
}

.modal h3 {
  margin-bottom: 24px;
  font-size: 20px;
  color: #00d4ff;
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
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
  color: #94a3b8;
  letter-spacing: 0.3px;
}

.form-item input {
  width: 100%;
  padding: 12px 16px;
  background: #0a0e27;
  border: 1px solid #2d3561;
  border-radius: 8px;
  font-size: 14px;
  color: #e0e7ff;
  box-sizing: border-box;
  transition: all 0.3s ease;
}

.form-item input::placeholder {
  color: #64748b;
}

.form-item input:focus {
  outline: none;
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}

.form-item input:disabled {
  background-color: rgba(10, 14, 39, 0.5);
  color: #64748b;
  cursor: not-allowed;
  border-color: #1e2642;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 10px 20px;
  background: transparent;
  color: #94a3b8;
  border: 1px solid #2d3561;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-cancel:hover {
  color: #00d4ff;
  border-color: #00d4ff;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.2);
}

.btn-confirm {
  padding: 10px 24px;
  background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
  color: #0a0e27;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
}

.btn-confirm:hover {
  box-shadow: 0 0 25px rgba(0, 212, 255, 0.5);
  transform: translateY(-2px);
}

.btn-confirm:disabled {
  background: linear-gradient(135deg, #3d4571 0%, #2d3561 100%);
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}
</style>
