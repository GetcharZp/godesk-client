<template>
  <div class="file-manager">
    <div class="connection-panel" v-if="!connectedDevice">
      <a-card title="连接远程设备" class="connection-card">
        <a-form layout="vertical">
          <a-form-item label="设备码">
            <a-input v-model:value="deviceCode" placeholder="请输入远程设备码" />
          </a-form-item>
          <a-form-item label="密码">
            <a-input-password v-model:value="devicePassword" placeholder="请输入连接密码" />
          </a-form-item>
          <a-form-item>
            <a-button type="primary" @click="connectToDevice" :loading="isConnecting" block>
              连接
            </a-button>
          </a-form-item>
        </a-form>
      </a-card>
    </div>

    <div class="file-panel" v-else>
      <div class="panel-header">
        <span class="device-info">
          <FolderOutlined /> 远程设备: {{ connectedDevice.code }}
        </span>
        <a-button size="small" @click="disconnect">断开连接</a-button>
      </div>

      <div class="file-container">
        <div class="local-files">
          <div class="section-header">
            <span><FolderOutlined /> 本地文件</span>
          </div>
          <div class="path-bar">
            <a-button size="small" @click="goToLocalParent" :disabled="localPath === ''">
              <ArrowUpOutlined />
            </a-button>
            <a-input v-model:value="localPath" readonly class="path-input" />
            <a-button size="small" @click="refreshLocalFiles">
              <ReloadOutlined />
            </a-button>
          </div>
          <div class="file-list">
            <div v-if="localLoading" class="loading">
              <a-spin />
            </div>
            <div v-else-if="localFiles.length === 0" class="empty">
              空文件夹
            </div>
            <div v-else>
              <div
                v-for="file in localFiles"
                :key="file.path"
                class="file-item"
                :class="{ selected: selectedLocalFile?.path === file.path }"
                @click="selectLocalFile(file)"
                @dblclick="openLocalFolder(file)"
              >
                <FolderOutlined v-if="file.isDir" class="file-icon folder" />
                <FileOutlined v-else class="file-icon" />
                <span class="file-name">{{ file.name }}</span>
                <span class="file-size" v-if="!file.isDir">{{ formatSize(file.size) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="transfer-buttons">
          <a-button type="primary" :disabled="!selectedLocalFile || selectedLocalFile.isDir || transferring" @click="uploadFile">
            <ArrowRightOutlined /> 上传
          </a-button>
          <a-button :disabled="!selectedRemoteFile || selectedRemoteFile.isDir || transferring" @click="downloadFile">
            <ArrowLeftOutlined /> 下载
          </a-button>
        </div>

        <div class="remote-files">
          <div class="section-header">
            <span><FolderOutlined /> 远程文件</span>
          </div>
          <div class="path-bar">
            <a-button size="small" @click="goToRemoteParent" :disabled="remotePath === ''">
              <ArrowUpOutlined />
            </a-button>
            <a-input v-model:value="remotePath" readonly class="path-input" />
            <a-button size="small" @click="refreshRemoteFiles">
              <ReloadOutlined />
            </a-button>
          </div>
          <div class="file-list">
            <div v-if="remoteLoading" class="loading">
              <a-spin />
            </div>
            <div v-else-if="remoteFiles.length === 0" class="empty">
              空文件夹
            </div>
            <div v-else>
              <div
                v-for="file in remoteFiles"
                :key="file.path"
                class="file-item"
                :class="{ selected: selectedRemoteFile?.path === file.path }"
                @click="selectRemoteFile(file)"
                @dblclick="openRemoteFolder(file)"
              >
                <FolderOutlined v-if="file.isDir" class="file-icon folder" />
                <FileOutlined v-else class="file-icon" />
                <span class="file-name">{{ file.name }}</span>
                <span class="file-size" v-if="!file.isDir">{{ formatSize(file.size) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="transfer-progress" v-if="transferring">
        <div class="progress-header">
          <span>{{ transferDirection === 'upload' ? '上传中' : '下载中' }}: {{ transferFileName }}</span>
          <a-button size="small" danger @click="cancelTransfer">取消</a-button>
        </div>
        <a-progress :percent="transferPercent" :status="transferStatus" />
        <div class="progress-info">
          {{ formatSize(transferReceived) }} / {{ formatSize(transferTotal) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  FolderOutlined,
  FileOutlined,
  ArrowUpOutlined,
  ArrowRightOutlined,
  ArrowLeftOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const deviceCode = ref('')
const devicePassword = ref('')
const isConnecting = ref(false)
const connectedDevice = ref(null)

const localPath = ref('')
const localFiles = ref([])
const localLoading = ref(false)
const selectedLocalFile = ref(null)

const remotePath = ref('')
const remoteFiles = ref([])
const remoteLoading = ref(false)
const selectedRemoteFile = ref(null)

const transferring = ref(false)
const transferDirection = ref('')
const transferFileName = ref('')
const transferPercent = ref(0)
const transferReceived = ref(0)
const transferTotal = ref(0)
const transferStatus = ref('active')
const transferId = ref('')

let pollTimer = null
let progressTimer = null

onMounted(async () => {
  if (route.query.deviceCode) {
    deviceCode.value = route.query.deviceCode
  }
  if (route.query.password) {
    devicePassword.value = route.query.password
  }
  if (route.query.sessionId && route.query.targetUUID) {
    connectedDevice.value = {
      code: route.query.deviceCode || '',
      sessionId: route.query.sessionId,
      uuid: route.query.targetUUID
    }
    fetchLocalFileList()
    fetchRemoteFileList()
  } else if (deviceCode.value && devicePassword.value) {
    await connectToDevice()
  }
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
  if (progressTimer) {
    clearInterval(progressTimer)
  }
})

const connectToDevice = async () => {
  if (!deviceCode.value.trim() || !devicePassword.value.trim()) {
    message.error('请输入设备码和密码')
    return
  }

  isConnecting.value = true
  try {
    const deviceCodeNum = parseInt(deviceCode.value)
    console.log('[FileManager] Connecting to device:', deviceCodeNum)

    // 检查是否已有该设备的文件会话
    const existingSessionRes = await getFileSessionByDeviceCode(deviceCodeNum)
    console.log('[FileManager] Existing file session:', existingSessionRes)
    
    if (existingSessionRes.code === 200 && existingSessionRes.data) {
      const existingSession = existingSessionRes.data
      if (existingSession.targetUUID || existingSession.TargetUUID || existingSession.targetUuid) {
        connectedDevice.value = {
          code: deviceCode.value,
          remark: '远程设备 ' + deviceCode.value,
          password: devicePassword.value,
          uuid: existingSession.targetUUID || existingSession.TargetUUID || existingSession.targetUuid,
          sessionId: existingSession.sessionId || existingSession.SessionId || existingSession.sessionId
        }
        message.success('连接成功（使用现有会话）')
        fetchLocalFileList()
        fetchRemoteFileList()
        return
      }
    }

    const res = await sendControlRequest({
      targetDeviceCode: deviceCodeNum,
      targetPassword: devicePassword.value,
      requestControl: false
    })

    console.log('[FileManager] Control request response:', res)

    if (!res || res.code !== 200 || !res.data) {
      throw new Error(res?.msg || '连接被拒绝，请检查设备码和密码')
    }

    let targetUUID = ''
    let sessionId = typeof res.data === 'string' ? res.data : (res.data.sessionId || res.data.SessionId)
    let retries = 0
    const maxRetries = 30

    while (!targetUUID && retries < maxRetries) {
      await new Promise(resolve => setTimeout(resolve, 100))

      const sessionRes = await getFileSessionByDeviceCode(deviceCodeNum)
      console.log('[FileManager] Get session retry ' + retries + ':', sessionRes)

      if (sessionRes.code === 200 && sessionRes.data) {
        targetUUID = sessionRes.data.TargetUUID || sessionRes.data.targetUUID || sessionRes.data.targetUuid || ''
        sessionId = sessionRes.data.SessionId || sessionRes.data.sessionId || sessionId
      }
      retries++
    }

    if (!targetUUID) {
      throw new Error('无法获取远程设备 UUID，设备可能未响应')
    }

    connectedDevice.value = {
      code: deviceCode.value,
      remark: '远程设备 ' + deviceCode.value,
      password: devicePassword.value,
      uuid: targetUUID,
      sessionId: sessionId
    }

    message.success('连接成功')
    fetchLocalFileList()
    fetchRemoteFileList()
  } catch (error) {
    message.error('连接失败: ' + error.message)
    console.error('Connect error:', error)
  } finally {
    isConnecting.value = false
  }
}

const disconnect = async () => {
  // 发送断开连接通知
  if (connectedDevice.value?.uuid) {
    try {
      await disconnectControl(connectedDevice.value.uuid)
    } catch (e) {
      console.error('Disconnect control error:', e)
    }
  }
  
  // 移除会话
  if (connectedDevice.value?.sessionId) {
    try {
      await removeSession(connectedDevice.value.sessionId)
    } catch (e) {
      console.error('Remove session error:', e)
    }
  }
  
  connectedDevice.value = null
  localFiles.value = []
  remoteFiles.value = []
  localPath.value = ''
  remotePath.value = ''
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  router.push('/remote-control')
}

const fetchLocalFileList = async () => {
  localLoading.value = true
  try {
    const res = await listLocalFiles(localPath.value)
    if (res.code === 200) {
      localFiles.value = res.data || []
      if (localPath.value === '' && res.data && res.data.length > 0 && res.data[0].path && res.data[0].path.includes(':\\')) {
        localPath.value = res.data[0].path.substring(0, 2) + '\\'
      }
    } else {
      message.error('获取本地文件列表失败')
    }
  } catch (error) {
    console.error('Fetch local files error:', error)
    message.error('获取本地文件列表失败')
  } finally {
    localLoading.value = false
  }
}

const fetchRemoteFileList = async () => {
  if (!connectedDevice.value?.sessionId) {
    console.error('[FileManager] No sessionId, cannot fetch remote files')
    return
  }

  console.log('[FileManager] Fetching remote files, sessionId:', connectedDevice.value.sessionId, 'uuid:', connectedDevice.value.uuid, 'path:', remotePath.value)

  remoteLoading.value = true
  try {
    await requestRemoteFileList(connectedDevice.value.sessionId, remotePath.value)

    // 等待响应，最多重试10次
    let res = null
    for (let i = 0; i < 10; i++) {
      await new Promise(resolve => setTimeout(resolve, 200))
      res = await getRemoteFileList(connectedDevice.value.uuid, remotePath.value)
      console.log('[FileManager] Get remote file list retry ' + i + ':', res)
      if (res.code === 200 && res.data && res.data.files) {
        break
      }
    }

    if (res && res.code === 200 && res.data) {
      // 规范化字段名
      const files = (res.data.files || res.data.Files || []).map(f => ({
        name: f.name || f.Name,
        path: f.path || f.Path,
        size: f.size || f.Size || 0,
        isDir: f.isDir || f.IsDir || f.is_dir || false,
        modifyTime: f.modifyTime || f.ModifyTime || f.modify_time || 0,
        mode: f.mode || f.Mode || 0
      }))
      remoteFiles.value = files
      remotePath.value = res.data.currentPath || res.data.CurrentPath || remotePath.value
      console.log('[FileManager] Normalized remote files:', files)
    } else {
      message.error('获取远程文件列表失败')
    }
  } catch (error) {
    console.error('Fetch remote files error:', error)
    message.error('获取远程文件列表失败')
  } finally {
    remoteLoading.value = false
  }
}

const refreshLocalFiles = () => fetchLocalFileList()
const refreshRemoteFiles = () => fetchRemoteFileList()

const goToLocalParent = () => {
  if (!localPath.value) return
  const parts = localPath.value.replace(/\\/g, '/').split('/').filter(p => p)
  parts.pop()
  localPath.value = parts.length > 0 ? parts.join('/') : ''
  fetchLocalFileList()
}

const goToRemoteParent = () => {
  if (!remotePath.value) return
  const parts = remotePath.value.replace(/\\/g, '/').split('/').filter(p => p)
  parts.pop()
  remotePath.value = parts.length > 0 ? parts.join('/') : ''
  fetchRemoteFileList()
}

const selectLocalFile = (file) => {
  selectedLocalFile.value = file
}

const selectRemoteFile = (file) => {
  selectedRemoteFile.value = file
}

const openLocalFolder = (file) => {
  if (file.isDir) {
    localPath.value = file.path
    fetchLocalFileList()
  }
}

const openRemoteFolder = (file) => {
  if (file.isDir) {
    remotePath.value = file.path
    fetchRemoteFileList()
  }
}

const uploadFile = async () => {
  if (!selectedLocalFile.value || selectedLocalFile.value.isDir) return
  
  const targetPath = remotePath.value 
    ? `${remotePath.value}/${selectedLocalFile.value.name}`.replace(/\\/g, '/')
    : selectedLocalFile.value.name

  // 检查目标文件是否存在
  try {
    const checkRes = await checkFileExistsApi(targetPath)
    if (checkRes.code === 200 && checkRes.data?.exists) {
      const confirmed = confirm(`文件 "${selectedLocalFile.value.name}" 已存在于目标位置，是否覆盖？`)
      if (!confirmed) return
    }
  } catch (e) {
    console.error('Check file exists error:', e)
  }

  try {
    const res = await uploadFileApi(
      connectedDevice.value.sessionId,
      selectedLocalFile.value.path,
      targetPath
    )
    if (res.code === 200) {
      if (res.data?.complete) {
        message.success('上传完成: ' + selectedLocalFile.value.name)
        refreshRemoteFiles()
      } else {
        startProgressTracking(res.data.transferId, 'upload', selectedLocalFile.value.name, res.data.totalSize)
        message.info('上传已开始: ' + selectedLocalFile.value.name)
      }
    } else {
      message.error('上传失败: ' + (res.msg || '未知错误'))
    }
  } catch (error) {
    console.error('Upload error:', error)
    message.error('上传失败')
  }
}

const downloadFile = async () => {
  if (!selectedRemoteFile.value || selectedRemoteFile.value.isDir) return
  
  const localTargetPath = localPath.value
    ? `${localPath.value}\\${selectedRemoteFile.value.name}`
    : selectedRemoteFile.value.name

  // 检查本地文件是否存在
  try {
    const checkRes = await checkFileExistsApi(localTargetPath)
    if (checkRes.code === 200 && checkRes.data?.exists) {
      const confirmed = confirm(`文件 "${selectedRemoteFile.value.name}" 已存在于本地，是否覆盖？`)
      if (!confirmed) return
    }
  } catch (e) {
    console.error('Check file exists error:', e)
  }

  try {
    const res = await downloadFileApi(
      connectedDevice.value.sessionId,
      selectedRemoteFile.value.path,
      localTargetPath
    )
    if (res.code === 200) {
      if (res.data?.complete) {
        message.success('下载完成: ' + selectedRemoteFile.value.name)
        refreshLocalFiles()
      } else {
        startProgressTracking(res.data.transferId, 'download', selectedRemoteFile.value.name, res.data.totalSize)
        message.info('下载已开始: ' + selectedRemoteFile.value.name)
      }
    } else {
      message.error('下载失败: ' + (res.msg || '未知错误'))
    }
  } catch (error) {
    console.error('Download error:', error)
    message.error('下载失败')
  }
}

const startProgressTracking = (tid, direction, fileName, total) => {
  transferId.value = tid
  transferDirection.value = direction
  transferFileName.value = fileName
  transferTotal.value = total || 0
  transferReceived.value = 0
  transferPercent.value = 0
  transferStatus.value = 'active'
  transferring.value = true

  progressTimer = setInterval(async () => {
    try {
      const res = await getFileTransferStatusApi(transferId.value)
      if (res.code === 200 && res.data) {
        transferReceived.value = res.data.received || 0
        transferTotal.value = res.data.total || transferTotal.value
        if (transferTotal.value > 0) {
          transferPercent.value = Math.round((transferReceived.value / transferTotal.value) * 100)
        }
        if (res.data.complete) {
          transferStatus.value = res.data.error ? 'exception' : 'success'
          clearInterval(progressTimer)
          progressTimer = null
          setTimeout(() => {
            transferring.value = false
            if (transferDirection.value === 'upload') {
              refreshRemoteFiles()
            } else {
              refreshLocalFiles()
            }
            if (res.data.error) {
              message.error('传输失败: ' + res.data.error)
            } else {
              message.success('传输完成: ' + transferFileName.value)
            }
          }, 500)
        }
      }
    } catch (e) {
      console.error('Get transfer status error:', e)
    }
  }, 200)
}

const cancelTransfer = async () => {
  if (!transferId.value) return
  try {
    await cancelFileTransferApi(connectedDevice.value.sessionId, transferId.value, '用户取消')
    message.info('已取消传输')
  } catch (e) {
    console.error('Cancel transfer error:', e)
  }
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
  transferring.value = false
}

const formatSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB'
}

// API 调用
const getFileSessionByDeviceCode = (deviceCode) => {
  if (window.go?.main?.App?.GetFileSessionByDeviceCode) {
    return window.go.main.App.GetFileSessionByDeviceCode(deviceCode)
  }
  return Promise.resolve({ code: 200, data: null })
}

const removeSession = (sessionId) => {
  if (window.go?.main?.App?.RemoveSession) {
    return window.go.main.App.RemoveSession(sessionId)
  }
  return Promise.resolve({ code: 200 })
}

const disconnectControl = (targetUUID) => {
  if (window.go?.main?.App?.DisconnectControl) {
    return window.go.main.App.DisconnectControl(targetUUID)
  }
  return Promise.resolve({ code: 200 })
}

const sendControlRequest = (params) => {
  if (window.go?.main?.App?.SendControlRequest) {
    return window.go.main.App.SendControlRequest(
      params.targetDeviceCode,
      params.targetPassword,
      params.requestControl || false
    )
  }
  return Promise.resolve({ code: -1, msg: 'API 不可用' })
}

const listLocalFiles = (path) => {
  if (window.go?.main?.App?.ListLocalFiles) {
    return window.go.main.App.ListLocalFiles(path)
  }
  return Promise.resolve({ code: 200, data: [] })
}

const requestRemoteFileList = (sessionId, path) => {
  if (window.go?.main?.App?.RequestRemoteFileList) {
    return window.go.main.App.RequestRemoteFileList(sessionId, path)
  }
  return Promise.resolve({ code: -1 })
}

const getRemoteFileList = (targetUUID, path) => {
  if (window.go?.main?.App?.GetRemoteFileList) {
    return window.go.main.App.GetRemoteFileList(targetUUID, path)
  }
  return Promise.resolve({ code: -1 })
}

const uploadFileApi = (sessionId, localPath, remotePath) => {
  if (window.go?.main?.App?.UploadFile) {
    return window.go.main.App.UploadFile(sessionId, localPath, remotePath)
  }
  return Promise.resolve({ code: -1, msg: 'API 不可用' })
}

const downloadFileApi = (sessionId, remotePath, localPath) => {
  if (window.go?.main?.App?.DownloadFile) {
    return window.go.main.App.DownloadFile(sessionId, remotePath, localPath)
  }
  return Promise.resolve({ code: -1, msg: 'API 不可用' })
}

const checkFileExistsApi = (path) => {
  if (window.go?.main?.App?.CheckFileExists) {
    return window.go.main.App.CheckFileExists(path)
  }
  return Promise.resolve({ code: 200, data: { exists: false } })
}

const getFileTransferStatusApi = (transferId) => {
  if (window.go?.main?.App?.GetFileTransferStatus) {
    return window.go.main.App.GetFileTransferStatus(transferId)
  }
  return Promise.resolve({ code: -1 })
}

const cancelFileTransferApi = (sessionId, transferId, reason) => {
  if (window.go?.main?.App?.CancelFileTransfer) {
    return window.go.main.App.CancelFileTransfer(sessionId, transferId, reason)
  }
  return Promise.resolve({ code: -1 })
}
</script>

<style scoped>
.file-manager {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.connection-panel {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  padding: 20px;
}

.connection-card {
  width: 100%;
  max-width: 400px;
}

.file-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e8e8e8;
}

.device-info {
  font-weight: 500;
}

.file-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.transfer-progress {
  margin: 8px;
  padding: 12px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
}

.progress-info {
  text-align: right;
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.local-files,
.remote-files {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #e8e8e8;
  margin: 8px;
  border-radius: 4px;
  overflow: hidden;
}

.section-header {
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #e8e8e8;
  font-weight: 500;
}

.path-bar {
  display: flex;
  gap: 8px;
  padding: 8px;
  border-bottom: 1px solid #e8e8e8;
}

.path-input {
  flex: 1;
}

.file-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.loading,
.empty {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100px;
  color: #999;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
}

.file-item:hover {
  background: #f5f5f5;
}

.file-item.selected {
  background: #e6f7ff;
}

.file-icon {
  margin-right: 8px;
  color: #999;
}

.file-icon.folder {
  color: #faad14;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: #999;
  font-size: 12px;
  margin-left: 8px;
}

.transfer-buttons {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
  padding: 8px;
}
</style>
