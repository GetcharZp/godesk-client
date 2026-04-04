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
  color: #00d4ff;
  font-size: 24px;
  font-weight: 500;
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
  letter-spacing: 0.5px;
}

.add-btn {
  padding: 10px 20px;
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

.add-btn:hover {
  background: linear-gradient(135deg, #00e5ff 0%, #00b3e6 100%);
  box-shadow: 0 0 25px rgba(0, 212, 255, 0.5);
  transform: translateY(-2px);
}

.search-bar {
  margin-bottom: 24px;
}

.search-bar input {
  width: 100%;
  padding: 12px 16px;
  background: #0a0e27;
  border: 1px solid #2d3561;
  border-radius: 8px;
  font-size: 14px;
  color: #e0e7ff;
  transition: all 0.3s ease;
}

.search-bar input::placeholder {
  color: #64748b;
}

.search-bar input:hover,
.search-bar input:focus {
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
  outline: none;
}

.device-categories {
  background: linear-gradient(135deg, #151b3d 0%, #1a2040 100%);
  padding: 24px;
  border-radius: 12px;
  border: 1px solid #2d3561;
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.3);
}

.category {
  margin-bottom: 32px;
}

.category:last-child {
  margin-bottom: 0;
}

.category h3 {
  font-size: 16px;
  color: #00d4ff;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #2d3561;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
}

.device-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.device-item {
  padding: 16px 20px;
  background: rgba(10, 14, 39, 0.5);
  border: 1px solid #2d3561;
  border-radius: 8px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
}

.device-item:hover {
  background: rgba(0, 212, 255, 0.08);
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.15);
}

.device-code {
  font-family: 'Courier New', monospace;
  font-weight: bold;
  color: #00d4ff;
  min-width: 100px;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
}

.device-remark {
  flex: 1;
  color: #e0e7ff;
}

.device-os {
  color: #94a3b8;
  font-size: 12px;
  min-width: 80px;
}

.device-status {
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 12px;
  background: rgba(100, 116, 139, 0.2);
  color: #64748b;
  border: 1px solid transparent;
}

.device-status.online {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border-color: #10b981;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
}

.device-actions {
  display: flex;
  gap: 8px;
}

.btn-control {
  padding: 6px 14px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-control:hover {
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.5);
  transform: translateY(-1px);
}

.btn-file {
  padding: 6px 14px;
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-file:hover {
  box-shadow: 0 0 15px rgba(124, 58, 237, 0.5);
  transform: translateY(-1px);
}

.btn-edit {
  padding: 6px 14px;
  background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
  color: #0a0e27;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-edit:hover {
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
  transform: translateY(-1px);
}

.btn-delete {
  padding: 6px 14px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-delete:hover {
  box-shadow: 0 0 15px rgba(239, 68, 68, 0.5);
  transform: translateY(-1px);
}

.empty-tip {
  text-align: center;
  color: #64748b;
  padding: 60px 20px;
  font-size: 14px;
}
</style>
