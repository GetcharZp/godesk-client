<template>
  <div class="file-manager" @contextmenu.prevent>
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
          <div class="file-list" tabindex="0" @keydown="handleLocalKeyDown">
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
                :class="{ selected: selectedLocalFile?.path === file.path, editing: editingLocalFile?.path === file.path }"
                @click.stop="selectLocalFile(file)"
                @dblclick.stop="openLocalFolder(file)"
                @contextmenu.stop.prevent="showLocalContextMenu($event, file)"
              >
                <FolderOutlined v-if="file.isDir" class="file-icon folder" />
                <FileOutlined v-else class="file-icon" />
                <template v-if="editingLocalFile?.path === file.path">
                  <input
                    v-model="editingFileName"
                    class="file-name-edit"
                    @blur="finishLocalEdit"
                    @keyup.enter="finishLocalEdit"
                    @keyup.escape="cancelLocalEdit"
                    @click.stop
                    ref="localEditInput"
                  />
                </template>
                <template v-else>
                  <span class="file-name">{{ file.name }}</span>
                </template>
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
          <div class="file-list" tabindex="0" @keydown="handleRemoteKeyDown">
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
                :class="{ selected: selectedRemoteFile?.path === file.path, editing: editingRemoteFile?.path === file.path }"
                @click.stop="selectRemoteFile(file)"
                @dblclick.stop="openRemoteFolder(file)"
                @contextmenu.stop.prevent="showRemoteContextMenu($event, file)"
              >
                <FolderOutlined v-if="file.isDir" class="file-icon folder" />
                <FileOutlined v-else class="file-icon" />
                <template v-if="editingRemoteFile?.path === file.path">
                  <input
                    v-model="editingFileName"
                    class="file-name-edit"
                    @blur="finishRemoteEdit"
                    @keyup.enter="finishRemoteEdit"
                    @keyup.escape="cancelRemoteEdit"
                    @click.stop
                    ref="remoteEditInput"
                  />
                </template>
                <template v-else>
                  <span class="file-name">{{ file.name }}</span>
                </template>
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

    <div
      v-if="localContextMenuVisible"
      class="context-menu"
      :style="{ left: localContextMenuX + 'px', top: localContextMenuY + 'px' }"
    >
      <div class="context-menu-item" @click="startLocalRenameFromMenu">重命名</div>
    </div>

    <div
      v-if="remoteContextMenuVisible"
      class="context-menu"
      :style="{ left: remoteContextMenuX + 'px', top: remoteContextMenuY + 'px' }"
    >
      <div class="context-menu-item" @click="startRemoteRenameFromMenu">重命名</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
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

const localContextMenuVisible = ref(false)
const localContextMenuX = ref(0)
const localContextMenuY = ref(0)
const localContextMenuFile = ref(null)

const remoteContextMenuVisible = ref(false)
const remoteContextMenuX = ref(0)
const remoteContextMenuY = ref(0)
const remoteContextMenuFile = ref(null)

const editingLocalFile = ref(null)
const editingRemoteFile = ref(null)
const editingFileName = ref('')
const originalFileName = ref('')

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

  document.addEventListener('click', hideAllContextMenu)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
  if (progressTimer) {
    clearInterval(progressTimer)
  }
  document.removeEventListener('click', hideAllContextMenu)
})

const hideAllContextMenu = () => {
  localContextMenuVisible.value = false
  remoteContextMenuVisible.value = false
}

const handleLocalKeyDown = (e) => {
  if (e.key === 'F2' && selectedLocalFile.value && !editingLocalFile.value) {
    e.preventDefault()
    startEditingLocal()
  }
}

const handleRemoteKeyDown = (e) => {
  if (e.key === 'F2' && selectedRemoteFile.value && !editingRemoteFile.value) {
    e.preventDefault()
    startEditingRemote()
  }
}

const startEditingLocal = () => {
  if (!selectedLocalFile.value) return
  editingLocalFile.value = selectedLocalFile.value
  editingFileName.value = selectedLocalFile.value.name
  originalFileName.value = selectedLocalFile.value.name
  nextTick(() => {
    const inputs = document.querySelectorAll('.file-name-edit')
    if (inputs.length > 0) {
      inputs[0].focus()
      inputs[0].select()
    }
  })
}

const startEditingRemote = () => {
  if (!selectedRemoteFile.value) return
  editingRemoteFile.value = selectedRemoteFile.value
  editingFileName.value = selectedRemoteFile.value.name
  originalFileName.value = selectedRemoteFile.value.name
  nextTick(() => {
    const inputs = document.querySelectorAll('.file-name-edit')
    if (inputs.length > 0) {
      inputs[0].focus()
      inputs[0].select()
    }
  })
}

const connectToDevice = async () => {
  if (!deviceCode.value.trim() || !devicePassword.value.trim()) {
    message.error('请输入设备码和密码')
    return
  }

  isConnecting.value = true
  try {
    const deviceCodeNum = parseInt(deviceCode.value)
    console.log('[FileManager] Connecting to device:', deviceCodeNum)

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
  if (connectedDevice.value?.uuid) {
    try {
      await disconnectControl(connectedDevice.value.uuid)
    } catch (e) {
      console.error('Disconnect control error:', e)
    }
  }
  
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
  localPath.value = parts.length > 0 ? parts.join('\\') : ''
  fetchLocalFileList()
}

const goToRemoteParent = () => {
  if (!remotePath.value) return
  const parts = remotePath.value.replace(/\\/g, '/').split('/').filter(p => p)
  parts.pop()
  remotePath.value = parts.length > 0 ? parts.join('\\') : ''
  fetchRemoteFileList()
}

const selectLocalFile = (file) => {
  selectedLocalFile.value = file
  editingLocalFile.value = null
}

const selectRemoteFile = (file) => {
  selectedRemoteFile.value = file
  editingRemoteFile.value = null
}

const openLocalFolder = (file) => {
  if (file.isDir) {
    localPath.value = file.path
    selectedLocalFile.value = null
    fetchLocalFileList()
  }
}

const openRemoteFolder = (file) => {
  if (file.isDir) {
    remotePath.value = file.path
    selectedRemoteFile.value = null
    fetchRemoteFileList()
  }
}

const uploadFile = async () => {
  if (!selectedLocalFile.value || selectedLocalFile.value.isDir) return
  
  const targetPath = remotePath.value 
    ? `${remotePath.value}\\${selectedLocalFile.value.name}`
    : selectedLocalFile.value.name

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
  return Promise.resolve({ code: -1 })
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
  return Promise.resolve({ code: -1 })
}

const downloadFileApi = (sessionId, remotePath, localPath) => {
  if (window.go?.main?.App?.DownloadFile) {
    return window.go.main.App.DownloadFile(sessionId, remotePath, localPath)
  }
  return Promise.resolve({ code: -1 })
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

const renameLocalFileApi = (oldPath, newName) => {
  if (window.go?.main?.App?.RenameLocalFile) {
    return window.go.main.App.RenameLocalFile(oldPath, newName)
  }
  return Promise.resolve({ code: -1 })
}

const renameRemoteFileApi = (sessionId, oldPath, newName) => {
  if (window.go?.main?.App?.RenameRemoteFile) {
    return window.go.main.App.RenameRemoteFile(sessionId, oldPath, newName)
  }
  return Promise.resolve({ code: -1 })
}

const getFileRenameResultApi = (requestId) => {
  if (window.go?.main?.App?.GetFileRenameResult) {
    return window.go.main.App.GetFileRenameResult(requestId)
  }
  return Promise.resolve({ code: -1 })
}

const showLocalContextMenu = (e, file) => {
  e.preventDefault()
  e.stopPropagation()
  selectedLocalFile.value = file
  localContextMenuFile.value = file
  localContextMenuX.value = e.clientX
  localContextMenuY.value = e.clientY
  localContextMenuVisible.value = true
  remoteContextMenuVisible.value = false
}

const showRemoteContextMenu = (e, file) => {
  e.preventDefault()
  e.stopPropagation()
  selectedRemoteFile.value = file
  remoteContextMenuFile.value = file
  remoteContextMenuX.value = e.clientX
  remoteContextMenuY.value = e.clientY
  remoteContextMenuVisible.value = true
  localContextMenuVisible.value = false
}

const startLocalRenameFromMenu = () => {
  localContextMenuVisible.value = false
  if (localContextMenuFile.value) {
    selectedLocalFile.value = localContextMenuFile.value
    startEditingLocal()
  }
}

const startRemoteRenameFromMenu = () => {
  remoteContextMenuVisible.value = false
  if (remoteContextMenuFile.value) {
    selectedRemoteFile.value = remoteContextMenuFile.value
    startEditingRemote()
  }
}

const cancelLocalEdit = () => {
  editingLocalFile.value = null
  editingFileName.value = ''
}

const cancelRemoteEdit = () => {
  editingRemoteFile.value = null
  editingFileName.value = ''
}

const finishLocalEdit = async () => {
  if (!editingLocalFile.value) return

  const file = editingLocalFile.value
  const newName = editingFileName.value.trim()

  if (!newName || newName === originalFileName.value) {
    editingLocalFile.value = null
    return
  }

  try {
    const res = await renameLocalFileApi(file.path, newName)
    if (res.code === 200) {
      message.success('重命名成功')
      refreshLocalFiles()
    } else {
      message.error('重命名失败: ' + (res.msg || '未知错误'))
    }
  } catch (e) {
    console.error('Rename local file error:', e)
    message.error('重命名失败')
  }

  editingLocalFile.value = null
}

const finishRemoteEdit = async () => {
  if (!editingRemoteFile.value) return

  const file = editingRemoteFile.value
  const newName = editingFileName.value.trim()

  if (!newName || newName === originalFileName.value) {
    editingRemoteFile.value = null
    return
  }

  try {
    const res = await renameRemoteFileApi(connectedDevice.value.sessionId, file.path, newName)
    if (res.code === 200) {
      if (res.data?.code === 0) {
        message.success('重命名成功')
        refreshRemoteFiles()
      } else {
        const requestId = res.data?.requestId
        if (requestId) {
          const checkResult = setInterval(async () => {
            const result = await getFileRenameResultApi(requestId)
            if (result.code === 200 && result.data?.exists) {
              clearInterval(checkResult)
              if (result.data.code === 0) {
                message.success('重命名成功')
                refreshRemoteFiles()
              } else {
                message.error('重命名失败: ' + result.data.message)
              }
            }
          }, 200)
          setTimeout(() => clearInterval(checkResult), 10000)
        }
      }
    } else {
      message.error('重命名失败: ' + (res.msg || '未知错误'))
    }
  } catch (e) {
    console.error('Rename remote file error:', e)
    message.error('重命名失败')
  }

  editingRemoteFile.value = null
}
</script>

<style scoped>
.file-manager {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #0a0e27 0%, #151b3d 100%);
  color: #e0e7ff;
}

.connection-panel {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  padding: 20px;
  background: radial-gradient(ellipse at center, #151b3d 0%, #0a0e27 100%);
}

.connection-card {
  width: 100%;
  max-width: 400px;
  background: rgba(26, 32, 64, 0.8);
  border: 1px solid #2d3561;
  border-radius: 12px;
  box-shadow: 0 0 30px rgba(0, 212, 255, 0.1);
  backdrop-filter: blur(10px);
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
  padding: 16px 20px;
  background: linear-gradient(90deg, #151b3d 0%, #1a2040 100%);
  border-bottom: 1px solid #2d3561;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
}

.device-info {
  font-weight: 500;
  color: #00d4ff;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
  font-size: 14px;
  letter-spacing: 0.5px;
}

.file-container {
  flex: 1;
  display: flex;
  overflow: hidden;
  background: #0a0e27;
}

.transfer-progress {
  margin: 12px;
  padding: 16px;
  background: linear-gradient(135deg, #151b3d 0%, #1a2040 100%);
  border: 1px solid #2d3561;
  border-radius: 8px;
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.1);
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  color: #e0e7ff;
}

.progress-info {
  text-align: right;
  font-size: 12px;
  color: #94a3b8;
  margin-top: 8px;
}

.local-files,
.remote-files {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #2d3561;
  margin: 12px;
  border-radius: 8px;
  overflow: hidden;
  background: linear-gradient(180deg, #151b3d 0%, #0f1428 100%);
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.3);
}

.section-header {
  padding: 12px 16px;
  background: linear-gradient(90deg, #1a2040 0%, #151b3d 100%);
  border-bottom: 1px solid #2d3561;
  font-weight: 500;
  color: #00d4ff;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
  letter-spacing: 0.5px;
}

.path-bar {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid #2d3561;
  background: rgba(21, 27, 61, 0.5);
}

.path-input {
  flex: 1;
  background: #0a0e27;
  border: 1px solid #2d3561;
  color: #e0e7ff;
  border-radius: 6px;
}

.path-input :deep(input) {
  background: transparent !important;
  color: #e0e7ff !important;
}

.path-input :deep(input::placeholder) {
  color: #64748b !important;
}

.file-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  outline: none;
  background: #0a0e27;
}

.file-list::-webkit-scrollbar {
  width: 8px;
}

.file-list::-webkit-scrollbar-track {
  background: #0a0e27;
}

.file-list::-webkit-scrollbar-thumb {
  background: #2d3561;
  border-radius: 4px;
}

.file-list::-webkit-scrollbar-thumb:hover {
  background: #3d4571;
}

.loading,
.empty {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100px;
  color: #64748b;
  font-size: 14px;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  margin-bottom: 4px;
}

.file-item:hover {
  background: rgba(0, 212, 255, 0.08);
  border-color: #2d3561;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.1);
}

.file-item.selected {
  background: rgba(0, 212, 255, 0.15);
  border-color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}

.file-item.editing {
  background: rgba(124, 58, 237, 0.15);
  border-color: #7c3aed;
  box-shadow: 0 0 15px rgba(124, 58, 237, 0.2);
}

.file-icon {
  margin-right: 10px;
  color: #64748b;
  flex-shrink: 0;
  font-size: 16px;
}

.file-icon.folder {
  color: #00d4ff;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.4);
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #e0e7ff;
  font-size: 13px;
}

.file-name-edit {
  flex: 1;
  border: 2px solid #00d4ff;
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 13px;
  outline: none;
  background: #0a0e27;
  color: #e0e7ff;
  min-width: 0;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
}

.file-size {
  color: #64748b;
  font-size: 12px;
  margin-left: 10px;
  flex-shrink: 0;
}

.transfer-buttons {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  padding: 0 12px;
}

.context-menu {
  position: fixed;
  background: linear-gradient(135deg, #1a2040 0%, #151b3d 100%);
  border: 1px solid #2d3561;
  border-radius: 8px;
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.5), 0 0 20px rgba(0, 212, 255, 0.1);
  z-index: 1000;
  min-width: 140px;
  padding: 6px 0;
  backdrop-filter: blur(10px);
}

.context-menu-item {
  padding: 10px 18px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #e0e7ff;
  font-size: 13px;
}

.context-menu-item:hover {
  background: rgba(0, 212, 255, 0.15);
  color: #00d4ff;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.5);
}

:deep(.ant-btn) {
  background: linear-gradient(135deg, #1a2040 0%, #151b3d 100%);
  border: 1px solid #2d3561;
  color: #e0e7ff;
  transition: all 0.3s ease;
}

:deep(.ant-btn:hover) {
  background: linear-gradient(135deg, #2d3561 0%, #1a2040 100%);
  border-color: #00d4ff;
  color: #00d4ff;
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
}

:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
  border: none;
  color: #0a0e27;
  font-weight: 500;
}

:deep(.ant-btn-primary:hover) {
  background: linear-gradient(135deg, #00e5ff 0%, #00b3e6 100%);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.5);
}

:deep(.ant-btn-dangerous) {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: #fff;
}

:deep(.ant-btn-dangerous:hover) {
  background: linear-gradient(135deg, #f87171 0%, #ef4444 100%);
  box-shadow: 0 0 20px rgba(239, 68, 68, 0.5);
}

:deep(.ant-input) {
  background: #0a0e27;
  border: 1px solid #2d3561;
  color: #e0e7ff;
}

:deep(.ant-input:hover),
:deep(.ant-input:focus) {
  border-color: #00d4ff;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.2);
}

:deep(.ant-input-password) {
  background: #0a0e27;
}

:deep(.ant-input-password :deep(input)) {
  background: transparent !important;
}

:deep(.ant-card) {
  background: rgba(26, 32, 64, 0.8);
  border: 1px solid #2d3561;
  color: #e0e7ff;
}

:deep(.ant-card-head) {
  background: transparent;
  border-bottom: 1px solid #2d3561;
  color: #00d4ff;
}

:deep(.ant-card-head-title) {
  color: #00d4ff;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
}

:deep(.ant-form-item-label > label) {
  color: #94a3b8;
}

:deep(.ant-progress-bg) {
  background: linear-gradient(90deg, #00d4ff 0%, #7c3aed 100%);
}

:deep(.ant-progress-text) {
  color: #00d4ff;
}

:deep(.ant-spin-dot-item) {
  background-color: #00d4ff;
}
</style>
