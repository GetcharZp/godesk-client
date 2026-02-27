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
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background-color: white;
  padding: 24px;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
}

.modal h3 {
  margin-bottom: 20px;
  font-size: 18px;
  color: #333;
}

.form {
  margin-bottom: 24px;
}

.form-item {
  margin-bottom: 16px;
}

.form-item label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  color: #666;
}

.form-item input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-item input:focus {
  outline: none;
  border-color: #1890ff;
}

.form-item input:disabled {
  background-color: #f5f5f5;
  color: #999;
  cursor: not-allowed;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 8px 16px;
  background-color: #f5f5f5;
  color: #666;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-cancel:hover {
  color: #40a9ff;
  border-color: #40a9ff;
}

.btn-confirm {
  padding: 8px 16px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.btn-confirm:hover {
  background-color: #40a9ff;
}

.btn-confirm:disabled {
  background-color: #bae7ff;
  cursor: not-allowed;
}
</style>
