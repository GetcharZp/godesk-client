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
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  color: #333;
  margin-bottom: 20px;
  font-size: 20px;
}

.search-bar {
  margin-bottom: 20px;
}

.search-bar input {
  width: 100%;
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.device-categories {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
}

.category {
  margin-bottom: 25px;
}

.category:last-child {
  margin-bottom: 0;
}

.category h3 {
  font-size: 16px;
  color: #333;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.device-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.device-item {
  padding: 12px 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 15px;
}

.device-code {
  font-family: monospace;
  font-weight: bold;
  color: #1890ff;
  min-width: 100px;
}

.device-remark {
  flex: 1;
  color: #333;
}

.device-os {
  color: #666;
  font-size: 12px;
  min-width: 80px;
}

.device-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  background-color: #f5f5f5;
  color: #999;
}

.device-status.online {
  background-color: #f6ffed;
  color: #52c41a;
}

.device-actions {
  display: flex;
  gap: 8px;
}

.btn-control {
  padding: 5px 12px;
  background-color: #52c41a;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-control:hover {
  background-color: #73d13d;
}

.btn-file {
  padding: 5px 12px;
  background-color: #722ed1;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-file:hover {
  background-color: #9254de;
}

.btn-edit {
  padding: 5px 12px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-edit:hover {
  background-color: #40a9ff;
}

.btn-delete {
  padding: 5px 12px;
  background-color: #ff4d4f;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-delete:hover {
  background-color: #ff7875;
}

.empty-tip {
  text-align: center;
  color: #999;
  padding: 40px;
  font-size: 14px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.add-btn {
  padding: 8px 16px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.add-btn:hover {
  background-color: #40a9ff;
}
</style>
