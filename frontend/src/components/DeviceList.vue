<template>
  <div class="device-list">
    <div class="header">
      <h2 class="page-title">设备列表</h2>
      <button class="add-btn" @click="openAddModal">+ 添加设备</button>
    </div>

    <div class="search-bar">
      <input type="text" placeholder="搜索设备" v-model="searchKeyword">
    </div>

    <div class="device-categories">
      <div class="category">
        <h3>我的设备 ({{ filteredDeviceList.length }})</h3>
        <div class="device-items">
          <div v-for="device in filteredDeviceList" :key="device.uuid" class="device-item">
            <span class="device-code">{{ device.code }}</span>
            <span class="device-remark">{{ device.remark || '未命名设备' }}</span>
            <span class="device-os">{{ device.os }}</span>
            <span class="device-status" :class="{ online: device.online }">
              {{ device.online ? '在线' : '离线' }}
            </span>
            <div class="device-actions">
              <button class="btn-control" @click="handleRemoteControl(device)">远程控制</button>
              <button class="btn-file" @click="handleRemoteFile(device)">远程文件</button>
              <button class="btn-edit" @click="openEditModal(device)">编辑</button>
              <button class="btn-delete" @click="handleDelete(device)">删除</button>
            </div>
          </div>
          <div v-if="filteredDeviceList.length === 0" class="empty-tip">暂无设备</div>
        </div>
      </div>
    </div>

    <!-- 设备表单弹窗 -->
    <DeviceFormModal
      v-model:visible="modalVisible"
      :is-edit="isEdit"
      :initial-data="modalData"
      :loading="modalLoading"
      @confirm="handleModalConfirm"
      @cancel="handleModalCancel"
    />

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDeviceList, addDevice, editDevice, deleteDevice } from '../api/device.js'
import { message, Modal } from 'ant-design-vue'
import DeviceFormModal from './DeviceFormModal.vue'

const router = useRouter()
const deviceList = ref([])
const searchKeyword = ref('')

// 弹窗相关状态
const modalVisible = ref(false)
const isEdit = ref(false)
const modalLoading = ref(false)
const currentEditUuid = ref('')
const modalData = ref({
  code: '',
  password: '',
  remark: ''
})

const filteredDeviceList = computed(() => {
  if (!searchKeyword.value) return deviceList.value
  const keyword = searchKeyword.value.toLowerCase()
  return deviceList.value.filter(device =>
    (device.code?.toString().includes(keyword)) ||
    (device.remark?.toLowerCase().includes(keyword))
  )
})

const fetchDeviceList = async () => {
  const res = await getDeviceList()
  if (res && res.code === 200) {
    deviceList.value = res.data || []
  }
}

// 打开添加弹窗
const openAddModal = () => {
  isEdit.value = false
  currentEditUuid.value = ''
  modalData.value = {
    code: '',
    password: '',
    remark: ''
  }
  modalVisible.value = true
}

// 打开编辑弹窗
const openEditModal = (device) => {
  isEdit.value = true
  currentEditUuid.value = device.uuid
  modalData.value = {
    code: device.code,
    password: '',
    remark: device.remark || ''
  }
  modalVisible.value = true
}

// 远程控制
const handleRemoteControl = (device) => {
  if (!device.online) {
    message.warning('设备离线，无法远程控制')
    return
  }
  // 跳转到远程控制页面，传递设备信息（包括设备码和密码）
  router.push({
    path: '/remote-control',
    query: {
      targetCode: device.code,
      targetName: device.remark || device.code,
      targetPassword: device.password
    }
  })
}

// 远程文件
const handleRemoteFile = (device) => {
  if (!device.online) {
    message.warning('设备离线，无法访问远程文件')
    return
  }
  router.push({
    path: '/file-manager',
    query: {
      deviceCode: device.code,
      password: device.password
    }
  })
}

// 弹窗确认
const handleModalConfirm = async (formData) => {
  if (!formData.code) {
    message.error('请输入设备码')
    return
  }

  modalLoading.value = true
  try {
    let res
    if (isEdit.value) {
      // 编辑模式
      if (!formData.password) {
        message.error('请输入设备密码')
        return
      }
      res = await editDevice({
        uuid: currentEditUuid.value,
        code: formData.code,
        password: formData.password,
        remark: formData.remark
      })
      if (res && res.code === 200) {
        message.success('编辑设备成功')
        modalVisible.value = false
        await fetchDeviceList()
      }
    } else {
      // 添加模式
      if (!formData.password) {
        message.error('请输入设备密码')
        return
      }
      res = await addDevice(formData)
      if (res && res.code === 200) {
        message.success('添加设备成功')
        modalVisible.value = false
        await fetchDeviceList()
      }
    }
  } finally {
    modalLoading.value = false
  }
}

// 弹窗取消
const handleModalCancel = () => {
  // 可以在这里添加取消时的逻辑
}

// 删除设备
const handleDelete = async (device) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除设备 "${device.remark || device.code}" 吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      const res = await deleteDevice({ uuid: device.uuid })
      if (res && res.code === 200) {
        message.success('删除设备成功')
        await fetchDeviceList()
      }
    }
  })
}

onMounted(() => {
  fetchDeviceList()
})
</script>

<style scoped>
.device-list {
  max-width: 900px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  color: var(--text-accent);
  font-size: 20px;
  font-weight: 500;
  text-shadow: 0 0 15px var(--accent-primary-glow-strong);
  letter-spacing: 0.5px;
}

.add-btn {
  padding: 10px 20px;
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

.add-btn:hover {
  box-shadow: 0 0 25px var(--accent-primary-glow-strong);
  transform: translateY(-2px);
}

.search-bar {
  margin-bottom: 24px;
}

.search-bar input {
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.search-bar input::placeholder {
  color: var(--text-muted);
}

.search-bar input:hover,
.search-bar input:focus {
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
  outline: none;
}

.device-categories {
  background: var(--bg-card);
  padding: 24px;
  border-radius: 12px;
  border: 1px solid var(--border-primary);
  box-shadow: var(--shadow-glow);
}

.category {
  margin-bottom: 32px;
}

.category:last-child {
  margin-bottom: 0;
}

.category h3 {
  font-size: 16px;
  color: var(--text-accent);
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-primary);
  text-shadow: 0 0 8px var(--accent-primary-glow);
}

.device-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.device-item {
  padding: 16px 20px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
}

.device-item:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-active);
  box-shadow: 0 0 15px var(--accent-primary-glow);
}

.device-code {
  font-family: 'Courier New', monospace;
  font-weight: bold;
  color: var(--text-accent);
  min-width: 100px;
  text-shadow: 0 0 8px var(--accent-primary-glow);
}

.device-remark {
  flex: 1;
  color: var(--text-primary);
}

.device-os {
  color: var(--text-secondary);
  font-size: 12px;
  min-width: 80px;
}

.device-status {
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 12px;
  background: var(--bg-status-offline);
  color: var(--text-muted);
  border: 1px solid transparent;
}

.device-status.online {
  background: var(--success-bg);
  color: var(--success);
  border-color: var(--success);
  box-shadow: 0 0 10px var(--success-glow);
}

.device-actions {
  display: flex;
  gap: 8px;
}

.btn-control {
  padding: 6px 14px;
  background: linear-gradient(135deg, var(--success) 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-control:hover {
  box-shadow: 0 0 15px var(--success-glow);
  transform: translateY(-1px);
}

.btn-file {
  padding: 6px 14px;
  background: linear-gradient(135deg, var(--purple) 0%, var(--purple-dark) 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-file:hover {
  box-shadow: 0 0 15px var(--purple-glow);
  transform: translateY(-1px);
}

.btn-edit {
  padding: 6px 14px;
  background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-primary-dark) 100%);
  color: var(--text-on-accent);
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-edit:hover {
  box-shadow: 0 0 15px var(--accent-primary-glow-strong);
  transform: translateY(-1px);
}

.btn-delete {
  padding: 6px 14px;
  background: linear-gradient(135deg, var(--danger) 0%, var(--danger-dark) 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-delete:hover {
  box-shadow: 0 0 15px var(--danger-glow);
  transform: translateY(-1px);
}

.empty-tip {
  text-align: center;
  color: var(--text-muted);
  padding: 60px 20px;
  font-size: 14px;
}
</style>
