<template>
  <div class="remote-control-page" :class="{ 'has-session': activeSessions.length > 0, 'sidebar-collapsed': isSidebarCollapsed && activeSessions.length > 0 }">
    <div class="header">
      <h2 class="page-title">远程控制</h2>
    </div>

    <div class="content-wrapper">
      <div class="main-section" :class="{ centered: activeSessions.length === 0, collapsed: isSidebarCollapsed && activeSessions.length > 0 }">
        <!-- 收起/展开按钮 -->
        <button
          v-if="activeSessions.length > 0"
          class="sidebar-toggle"
          @click="isSidebarCollapsed = !isSidebarCollapsed"
          :title="isSidebarCollapsed ? '展开' : '收起'"
        >
          {{ isSidebarCollapsed ? '→' : '←' }}
        </button>
        <div class="section-card">
          <h3>本机设备</h3>
          <div class="device-info-row">
            <span class="info-label">设备码</span>
            <span class="info-value">{{ myDeviceInfo.code || '-' }}</span>
            <button class="btn-copy" @click="copyToClipboard(myDeviceInfo.code)" :disabled="!myDeviceInfo.code">
              复制
            </button>
          </div>
          <div class="device-info-row">
            <span class="info-label">密码</span>
            <span class="info-value password">{{ showPassword ? myDeviceInfo.password : '******' }}</span>
            <button class="btn-toggle" @click="showPassword = !showPassword">
              {{ showPassword ? '隐藏' : '显示' }}
            </button>
            <button class="btn-copy" @click="copyToClipboard(myDeviceInfo.password)" :disabled="!myDeviceInfo.password">
              复制
            </button>
          </div>
        </div>

        <div class="section-card">
          <h3>远程连接</h3>
          <div class="connect-form">
            <div class="form-row">
              <div class="form-item">
                <label>设备码</label>
                <input
                  type="text"
                  v-model="remoteDeviceCode"
                  placeholder="输入对方设备码"
                />
              </div>
              <div class="form-item">
                <label>密码</label>
                <input
                  :type="showRemotePassword ? 'text' : 'password'"
                  v-model="remotePassword"
                  placeholder="输入对方密码"
                />
                <button class="btn-toggle-pwd" @click="showRemotePassword = !showRemotePassword">
                  {{ showRemotePassword ? '隐藏' : '显示' }}
                </button>
              </div>
            </div>
            <div class="form-actions">
              <button class="btn-primary" @click="startRemoteControl" :disabled="connecting">
                {{ connecting ? '连接中...' : '远程控制' }}
              </button>
              <button class="btn-secondary" @click="startRemoteFile" :disabled="connecting">
                远程文件
              </button>
            </div>
          </div>
        </div>

        <div class="section-card sessions-card" v-if="activeSessions.length > 0">
          <h3>活跃会话 ({{ activeSessions.length }})</h3>
          <div class="session-list">
            <div
              v-for="session in activeSessions"
              :key="session.sessionId"
              class="session-item"
              :class="{ active: currentSessionId === session.sessionId }"
              @click="selectSession(session.sessionId)"
            >
              <span class="session-code">{{ session.deviceName }}</span>
              <span class="session-status" :class="session.status">
                {{ getStatusText(session.status) }}
              </span>
              <div class="session-actions">
                <button class="btn-view" @click.stop="selectSession(session.sessionId)">查看</button>
                <button class="btn-disconnect" @click.stop="closeSession(session.sessionId)">断开</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="screen-section" v-if="activeSessions.length > 0">
        <div class="screen-header">
          <div class="session-tabs">
            <div
              v-for="session in activeSessions"
              :key="session.sessionId"
              class="session-tab"
              :class="{ active: currentSessionId === session.sessionId }"
              @click="selectSession(session.sessionId)"
            >
              <span class="tab-name">{{ session.deviceName }}</span>
              <span class="tab-status" :class="session.status"></span>
              <button class="tab-close" @click.stop="closeSession(session.sessionId)">×</button>
            </div>
          </div>
          <div class="screen-toolbar" v-if="currentSession">
            <button class="toolbar-btn" :class="{ active: currentSession.viewOnly }" @click="toggleViewOnly">
              仅查看
            </button>
            <button class="toolbar-btn" @click="toggleFullscreen">
              {{ isFullscreen ? '退出全屏' : '全屏' }}
            </button>
            <button class="toolbar-btn" @click="refreshScreen">刷新</button>
            <button class="toolbar-btn danger" @click="disconnectCurrent">断开</button>
          </div>
        </div>
        <div class="screen-wrapper" ref="screenWrapper">
          <template v-if="currentSession">
            <canvas
              ref="screenCanvas"
              class="screen-canvas"
              :width="currentSession.screenWidth || 1920"
              :height="currentSession.screenHeight || 1080"
              @mousemove="handleMouseMove"
              @mousedown="handleMouseDown"
              @mouseup="handleMouseUp"
              @wheel="handleMouseWheel"
              @keydown="handleKeyDown"
              @keyup="handleKeyUp"
              tabindex="0"
              :class="{ 'control-mode': !currentSession.viewOnly && currentSession.status === 'connected' }"
            ></canvas>
            <div v-if="currentSession.status === 'connecting'" class="screen-overlay">
              <div class="loading-spinner"></div>
              <p>正在连接...</p>
            </div>
            <div v-if="currentSession.status === 'error'" class="screen-overlay error">
              <p>连接失败</p>
              <button class="btn-retry" @click="reconnect">重新连接</button>
            </div>
          </template>
          <div v-else class="screen-overlay">
            <p>请选择一个会话</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { sendControlRequest, disconnectControl, sendMouseMove, sendMouseClick, sendMouseScroll, sendKeyDown, sendKeyUp } from '../api/channel.js'
import { getDeviceInfo, getDeviceList } from '../api/device.js'
import { getAllSessions, createSession, removeSession, getSessionByDeviceCode } from '../api/session.js'
import { startScreenStream } from '../api/screen.js'

const route = useRoute()
const router = useRouter()

const myDeviceInfo = ref({ code: '', password: '' })
const showPassword = ref(false)

const remoteDeviceCode = ref('')
const remotePassword = ref('')
const showRemotePassword = ref(false)
const connecting = ref(false)

const activeSessions = ref([])
const currentSessionId = ref(null)
const screenCanvas = ref(null)
const screenWrapper = ref(null)
const isFullscreen = ref(false)
const deviceList = ref([])
const screenStreamStopFns = ref(new Map()) // sessionId -> stopFunction
const isSidebarCollapsed = ref(false) // 左侧导航栏是否收起

const currentSession = computed(() => {
  return activeSessions.value.find(s => s.sessionId === currentSessionId.value)
})

// 获取设备显示名称（优先显示备注）
const getDeviceDisplayName = (deviceCode) => {
  const device = deviceList.value.find(d => d.code === deviceCode)
  if (device && device.remark) {
    return device.remark
  }
  return String(deviceCode)
}

const getStatusText = (status) => {
  const statusMap = {
    'connecting': '连接中',
    'connected': '已连接',
    'error': '失败',
    'disconnected': '已断开'
  }
  return statusMap[status] || status
}

const copyToClipboard = (text) => {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => {
    message.success('已复制')
  }).catch(() => {
    message.error('复制失败')
  })
}

const fetchMyDeviceInfo = async () => {
  try {
    const res = await getDeviceInfo()
    if (res && res.code === 200 && res.data) {
      myDeviceInfo.value = {
        code: res.data.code || '',
        password: res.data.password || ''
      }
    }
  } catch (error) {
    console.error('获取设备信息失败:', error)
  }
}

const fetchDeviceList = async () => {
  try {
    const res = await getDeviceList()
    if (res && res.code === 200 && res.data) {
      deviceList.value = res.data || []
    }
  } catch (error) {
    console.error('获取设备列表失败:', error)
  }
}

const loadSessions = async () => {
  try {
    const res = await getAllSessions()
    if (res && res.code === 200 && res.data) {
      // 为每个会话设置 deviceName
      activeSessions.value = (res.data || []).map(sess => ({
        ...sess,
        deviceName: getDeviceDisplayName(sess.deviceCode)
      }))
    }
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

const findSessionByDeviceCode = async (deviceCode) => {
  try {
    const res = await getSessionByDeviceCode(deviceCode)
    if (res && res.code === 200 && res.data) {
      return res.data
    }
  } catch (error) {
    console.error('查找会话失败:', error)
  }
  return null
}

const doConnect = async (deviceCode, password) => {
  const existingSession = await findSessionByDeviceCode(deviceCode)
  if (existingSession) {
    await loadSessions()
    currentSessionId.value = existingSession.sessionId
    message.info('该设备已连接，已切换到该会话')
    return true
  }

  connecting.value = true
  try {
    const res = await sendControlRequest({
      targetDeviceCode: deviceCode,
      targetPassword: password,
      requestControl: true
    })

    console.log('sendControlRequest response:', res)

    if (res && res.code === 200 && res.data) {
      const sessionId = res.data
      
      // 获取设备显示名称（优先显示备注）
      const deviceName = getDeviceDisplayName(deviceCode)

      // 先添加一个临时会话到列表中
      const tempSession = {
        sessionId: sessionId,
        deviceCode: deviceCode,
        deviceName: deviceName,
        viewOnly: false,
        status: 'connecting',
        screenWidth: res.data.targetInfo?.screenWidth || 1920,
        screenHeight: res.data.targetInfo?.screenHeight || 1080,
        createdAt: Date.now() / 1000,
        updatedAt: Date.now() / 1000
      }
      
      // 检查是否已存在（按 deviceCode 去重）
      const existingIndex = activeSessions.value.findIndex(s => s.deviceCode === deviceCode)
      if (existingIndex === -1) {
        activeSessions.value.push(tempSession)
      } else {
        // 如果已存在，更新现有会话的 sessionId 和名称
        activeSessions.value[existingIndex].sessionId = sessionId
        activeSessions.value[existingIndex].deviceName = deviceName
        activeSessions.value[existingIndex].status = 'connecting'
      }
      
      // 立即设置当前会话ID
      currentSessionId.value = sessionId
      
      // 等待 Vue 更新计算属性
      await nextTick()
      
      console.log('After setting currentSessionId:', currentSessionId.value, 'currentSession:', currentSession.value)
      
      // 然后从后端加载完整数据（但不覆盖当前选中的会话）
      const res2 = await getAllSessions()
      if (res2 && res2.code === 200 && res2.data) {
        // 合并后端数据，保留当前选中的会话
        const backendSessions = res2.data || []
        backendSessions.forEach(backendSess => {
          const idx = activeSessions.value.findIndex(s => s.deviceCode === backendSess.deviceCode)
          if (idx !== -1) {
            // 更新现有会话，但保留前端的 deviceName
            const existingDeviceName = activeSessions.value[idx].deviceName
            activeSessions.value[idx] = { ...activeSessions.value[idx], ...backendSess }
            activeSessions.value[idx].deviceName = existingDeviceName
          } else {
            // 添加新会话，设置 deviceName
            backendSess.deviceName = getDeviceDisplayName(backendSess.deviceCode)
            activeSessions.value.push(backendSess)
          }
        })
      }
      
      message.success('连接成功')
      return true
    } else {
      message.error(res?.msg || '连接被拒绝')
      return false
    }
  } catch (error) {
    message.error('连接失败: ' + error.message)
    return false
  } finally {
    connecting.value = false
  }
}

const startRemoteControl = async () => {
  if (!remoteDeviceCode.value) {
    message.error('请输入设备码')
    return
  }
  if (!remotePassword.value) {
    message.error('请输入密码')
    return
  }

  const deviceCode = parseInt(remoteDeviceCode.value)
  const success = await doConnect(deviceCode, remotePassword.value)
  
  if (success) {
    remoteDeviceCode.value = ''
    remotePassword.value = ''
  }
}

const startRemoteFile = () => {
  if (!remoteDeviceCode.value || !remotePassword.value) {
    message.error('请输入设备码和密码')
    return
  }
  message.info('远程文件功能开发中...')
}

const selectSession = (sessionId) => {
  currentSessionId.value = sessionId
  // 启动该会话的屏幕流
  startSessionScreenStream(sessionId)
}

// 启动会话的屏幕流
const startSessionScreenStream = (sessionId) => {
  // 如果已经在接收，先停止
  if (screenStreamStopFns.value.has(sessionId)) {
    return // 已经在接收
  }

  // 获取 canvas 元素用于视频解码渲染
  const canvas = currentSessionId.value === sessionId ? screenCanvas.value : null

  const stopFn = startScreenStream(sessionId, (imageUrl, data) => {
    // 更新会话的图像数据
    const session = activeSessions.value.find(s => s.sessionId === sessionId)
    if (session) {
      // 根据数据类型处理
      if (data.codec === 'h264' || data.codec === 'h265') {
        // 视频流格式
        session.screenWidth = data.width || session.screenWidth
        session.screenHeight = data.height || session.screenHeight
        session.codec = data.codec
        session.sequence = data.sequence
        
        // 收到屏幕数据，更新状态为已连接
        if (session.status === 'connecting') {
          session.status = 'connected'
        }
        
        // 视频帧已经通过解码器渲染到 canvas，无需额外处理
        if (data.decoded) {
          // 解码成功
        } else if (data.error) {
          console.warn('Video decode error:', data.error)
        }
      } else {
        // JPEG 图像格式
        session.lastImageUrl = imageUrl
        session.screenWidth = data.width || session.screenWidth
        session.screenHeight = data.height || session.screenHeight
        session.codec = data.codec || 'jpeg'
        session.sequence = data.sequence
        
        // 收到屏幕数据，更新状态为已连接
        if (session.status === 'connecting') {
          session.status = 'connected'
        }
        
        // 如果是当前选中的会话，渲染到 canvas
        if (currentSessionId.value === sessionId && screenCanvas.value && imageUrl) {
          renderImageToCanvas(imageUrl)
        }
      }
    }
  }, canvas)

  screenStreamStopFns.value.set(sessionId, stopFn)
}

// 停止会话的屏幕流
const stopSessionScreenStream = (sessionId) => {
  const stopFn = screenStreamStopFns.value.get(sessionId)
  if (stopFn) {
    stopFn()
    screenStreamStopFns.value.delete(sessionId)
  }
}

// 渲染图像到 canvas
const renderImageToCanvas = (imageUrl) => {
  if (!screenCanvas.value) return

  const ctx = screenCanvas.value.getContext('2d')
  const img = new Image()
  img.onload = () => {
    const canvas = screenCanvas.value
    const container = canvas.parentElement

    // 设置 canvas 为原始图像分辨率（保持清晰）
    canvas.width = img.width
    canvas.height = img.height

    // 使用 CSS 缩放来适应容器
    if (container) {
      const scale = Math.min(
        container.clientWidth / img.width,
        container.clientHeight / img.height,
        1
      )
      canvas.style.width = `${img.width * scale}px`
      canvas.style.height = `${img.height * scale}px`
    }

    // 使用高质量缩放
    ctx.imageSmoothingEnabled = true
    ctx.imageSmoothingQuality = 'high'
    ctx.drawImage(img, 0, 0)
  }
  img.src = imageUrl
}

const closeSession = async (sessionId) => {
  const session = activeSessions.value.find(s => s.sessionId === sessionId)
  if (session) {
    try {
      await disconnectControl({
        sessionId: sessionId,
        targetDeviceCode: session.deviceCode
      })
    } catch (error) {
      console.error('断开连接失败:', error)
    }
  }

  // 停止屏幕流
  stopSessionScreenStream(sessionId)
  await removeSession(sessionId)
  await loadSessions()

  if (currentSessionId.value === sessionId) {
    currentSessionId.value = activeSessions.value.length > 0 ? activeSessions.value[0].sessionId : null
  }
  message.success('已断开连接')
}

const toggleViewOnly = () => {
  if (currentSession.value) {
    currentSession.value.viewOnly = !currentSession.value.viewOnly
    message.info(currentSession.value.viewOnly ? '仅查看模式' : '控制模式')
  }
}

const toggleFullscreen = () => {
  if (!screenWrapper.value) return

  if (!document.fullscreenElement) {
    screenWrapper.value.requestFullscreen().then(() => {
      isFullscreen.value = true
    }).catch(err => {
      message.error('无法进入全屏模式')
    })
  } else {
    document.exitFullscreen().then(() => {
      isFullscreen.value = false
    })
  }
}

const refreshScreen = () => {
  message.info('刷新屏幕')
}

const disconnectCurrent = () => {
  if (currentSessionId.value) {
    closeSession(currentSessionId.value)
  }
}

const reconnect = () => {
  if (currentSession.value) {
    currentSession.value.status = 'connecting'
  }
}

const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

// 将 canvas 坐标转换为原始屏幕坐标
const convertToScreenCoordinates = (clientX, clientY) => {
  if (!screenCanvas.value || !currentSession.value) return { x: 0, y: 0 }

  const canvas = screenCanvas.value
  const rect = canvas.getBoundingClientRect()

  // 计算在 canvas 中的相对位置（0-1 之间）
  const relativeX = (clientX - rect.left) / rect.width
  const relativeY = (clientY - rect.top) / rect.height

  // 转换为原始屏幕坐标
  const screenX = Math.round(relativeX * currentSession.value.screenWidth)
  const screenY = Math.round(relativeY * currentSession.value.screenHeight)

  return { x: screenX, y: screenY }
}

// 鼠标移动事件处理
const handleMouseMove = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  const { x, y } = convertToScreenCoordinates(e.clientX, e.clientY)
  sendMouseMove(currentSession.value.sessionId, x, y)
}

// 鼠标按下事件处理
const handleMouseDown = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  const { x, y } = convertToScreenCoordinates(e.clientX, e.clientY)
  const button = e.button // 0=左键, 1=中键, 2=右键
  sendMouseClick(currentSession.value.sessionId, x, y, button, 'down')
}

// 鼠标释放事件处理
const handleMouseUp = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  const { x, y } = convertToScreenCoordinates(e.clientX, e.clientY)
  const button = e.button
  sendMouseClick(currentSession.value.sessionId, x, y, button, 'up')
}

// 鼠标滚轮事件处理
const handleMouseWheel = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  e.preventDefault()
  const { x, y } = convertToScreenCoordinates(e.clientX, e.clientY)
  sendMouseScroll(currentSession.value.sessionId, x, y, e.deltaX, e.deltaY)
}

// 键盘映射表：将 JavaScript key 转换为 robotgo 支持的键名
const keyMapping = {
  // 字母键
  'a': 'a', 'b': 'b', 'c': 'c', 'd': 'd', 'e': 'e', 'f': 'f', 'g': 'g', 'h': 'h',
  'i': 'i', 'j': 'j', 'k': 'k', 'l': 'l', 'm': 'm', 'n': 'n', 'o': 'o', 'p': 'p',
  'q': 'q', 'r': 'r', 's': 's', 't': 't', 'u': 'u', 'v': 'v', 'w': 'w', 'x': 'x',
  'y': 'y', 'z': 'z',
  // 数字键
  '0': '0', '1': '1', '2': '2', '3': '3', '4': '4',
  '5': '5', '6': '6', '7': '7', '8': '8', '9': '9',
  // 功能键
  'F1': 'f1', 'F2': 'f2', 'F3': 'f3', 'F4': 'f4', 'F5': 'f5',
  'F6': 'f6', 'F7': 'f7', 'F8': 'f8', 'F9': 'f9', 'F10': 'f10',
  'F11': 'f11', 'F12': 'f12',
  // 控制键
  'Enter': 'enter', 'Tab': 'tab', 'Backspace': 'backspace', 'Escape': 'esc',
  'Space': 'space', 'Delete': 'delete', 'Insert': 'insert',
  'Home': 'home', 'End': 'end', 'PageUp': 'pageup', 'PageDown': 'pagedown',
  // 方向键
  'ArrowUp': 'up', 'ArrowDown': 'down', 'ArrowLeft': 'left', 'ArrowRight': 'right',
  // 修饰键
  'Control': 'ctrl', 'Alt': 'alt', 'Shift': 'shift', 'Meta': 'cmd',
  // 特殊字符
  '-': '-', '=': '=', '[': '[', ']': ']', '\\': '\\', ';': ';', "'": "'",
  ',': ',', '.': '.', '/': '/', '`': '`'
}

// 获取修饰键列表
const getModifiers = (e) => {
  const modifiers = []
  if (e.ctrlKey) modifiers.push('ctrl')
  if (e.altKey) modifiers.push('alt')
  if (e.shiftKey) modifiers.push('shift')
  if (e.metaKey) modifiers.push('cmd')
  return modifiers
}

// 键盘按下事件处理
const handleKeyDown = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  // 阻止默认行为（如页面滚动）
  e.preventDefault()

  const key = keyMapping[e.key] || e.key.toLowerCase()
  if (!key) return

  const modifiers = getModifiers(e)
  sendKeyDown(currentSession.value.sessionId, key, modifiers)
}

// 键盘释放事件处理
const handleKeyUp = (e) => {
  if (!currentSession.value || currentSession.value.viewOnly || currentSession.value.status !== 'connected') return

  e.preventDefault()

  const key = keyMapping[e.key] || e.key.toLowerCase()
  if (!key) return

  const modifiers = getModifiers(e)
  sendKeyUp(currentSession.value.sessionId, key, modifiers)
}

const handleRouteQuery = async () => {
  const { targetCode, targetPassword } = route.query
  if (targetCode && targetPassword) {
    const deviceCode = parseInt(targetCode)
    const success = await doConnect(deviceCode, targetPassword)
    if (success) {
      // 清除路由参数，避免刷新时重复连接
      router.replace({ path: '/remote-control' })
    }
  }
}

onMounted(async () => {
  await fetchMyDeviceInfo()
  await fetchDeviceList() // 加载设备列表用于获取备注
  // 先处理路由参数（连接新设备），然后再加载会话列表
  // 这样可以确保新连接的会话被正确添加到列表中
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  await handleRouteQuery()
  // 最后再加载会话列表（包含新连接的会话）
  await loadSessions()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  // 停止所有屏幕流
  screenStreamStopFns.value.forEach((stopFn) => stopFn())
  screenStreamStopFns.value.clear()
})

// 监听当前会话变化，启动屏幕流
watch(currentSessionId, (newSessionId) => {
  if (newSessionId) {
    startSessionScreenStream(newSessionId)
  }
})
</script>

<style scoped>
.remote-control-page {
  max-width: 1200px;
  margin: 0 auto;
  height: calc(100vh - 40px);
  display: flex;
  flex-direction: column;
}

.remote-control-page.has-session {
  max-width: none;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.page-title {
  color: #333;
  margin: 0;
  font-size: 20px;
}

.content-wrapper {
  display: flex;
  gap: 20px;
  flex: 1;
  min-height: 0;
}

.main-section {
  width: 400px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-height: 100%;
  overflow-y: auto;
  position: relative;
  transition: width 0.3s ease;
}

.main-section.centered {
  width: 480px;
  margin: 0 auto;
}

.main-section.collapsed {
  width: 40px;
  overflow: hidden;
}

.main-section.collapsed .section-card,
.main-section.collapsed .sessions-card {
  display: none;
}

.sidebar-toggle {
  position: absolute;
  right: 8px;
  top: 8px;
  width: 24px;
  height: 24px;
  border: none;
  background-color: #f0f0f0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  transition: background-color 0.2s;
}

.sidebar-toggle:hover {
  background-color: #e0e0e0;
}

.main-section.collapsed .sidebar-toggle {
  right: 8px;
  top: 8px;
}

.sessions-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.sessions-card h3 {
  flex-shrink: 0;
}

.sessions-card .session-list {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.screen-section {
  flex: 1;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.screen-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: 1px solid #eee;
  background-color: #fafafa;
  flex-shrink: 0;
}

.session-tabs {
  display: flex;
  gap: 4px;
  overflow-x: auto;
  flex: 1;
}

.session-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #f0f0f0;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
}

.session-tab:hover {
  background-color: #e0e0e0;
}

.session-tab.active {
  background-color: #1890ff;
  color: white;
}

.tab-name {
  font-size: 13px;
}

.tab-status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #999;
}

.tab-status.connected {
  background-color: #52c41a;
}

.tab-status.connecting {
  background-color: #faad14;
}

.tab-status.error,
.tab-status.disconnected {
  background-color: #ff4d4f;
}

.tab-close {
  width: 14px;
  height: 14px;
  border: none;
  background: none;
  cursor: pointer;
  color: inherit;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  margin-left: 4px;
  opacity: 0.6;
}

.tab-close:hover {
  opacity: 1;
}

.screen-toolbar {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.toolbar-btn {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
}

.toolbar-btn:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.toolbar-btn.active {
  background-color: #1890ff;
  color: white;
  border-color: #1890ff;
}

.toolbar-btn.danger {
  color: #ff4d4f;
  border-color: #ffccc7;
}

.toolbar-btn.danger:hover {
  background-color: #ff4d4f;
  color: white;
}

.screen-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #1a1a1a;
  position: relative;
  overflow: auto;
  min-height: 0;
}

.screen-wrapper:-webkit-full-screen {
  background-color: #1a1a1a;
}

.screen-canvas {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.screen-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
}

.screen-overlay.error {
  color: #ff4d4f;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn-retry {
  margin-top: 12px;
  padding: 8px 20px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-retry:hover {
  background-color: #40a9ff;
}

.section-card {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.section-card h3 {
  font-size: 16px;
  color: #333;
  margin: 0 0 15px 0;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.device-info-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.device-info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-size: 14px;
  color: #666;
  min-width: 50px;
}

.info-value {
  flex: 1;
  font-family: 'Courier New', monospace;
  font-size: 16px;
  font-weight: 600;
  color: #333;
  background-color: #f5f5f5;
  padding: 8px 12px;
  border-radius: 4px;
}

.info-value.password {
  letter-spacing: 2px;
}

.btn-copy, .btn-toggle {
  padding: 6px 12px;
  font-size: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  color: #666;
  transition: all 0.3s;
}

.btn-copy:hover, .btn-toggle:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.btn-copy:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.connect-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-item {
  flex: 1;
  position: relative;
}

.form-item label {
  display: block;
  font-size: 14px;
  color: #666;
  margin-bottom: 6px;
}

.form-item input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-item input:focus {
  border-color: #1890ff;
  outline: none;
}

.btn-toggle-pwd {
  position: absolute;
  right: 8px;
  top: 28px;
  padding: 2px 8px;
  font-size: 12px;
  border: none;
  background: none;
  cursor: pointer;
  color: #999;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.btn-primary {
  flex: 1;
  padding: 10px 16px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-primary:hover:not(:disabled) {
  background-color: #40a9ff;
}

.btn-primary:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
}

.btn-secondary {
  flex: 1;
  padding: 10px 16px;
  background-color: white;
  color: #666;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-secondary:hover:not(:disabled) {
  border-color: #1890ff;
  color: #1890ff;
}

.session-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  border-left: 3px solid transparent;
}

.session-item:hover {
  background-color: #f0f0f0;
}

.session-item.active {
  border-left-color: #1890ff;
  background-color: #e6f7ff;
}

.session-code {
  flex: 1;
  font-size: 14px;
  color: #333;
}

.session-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.session-status.connected {
  background-color: #f6ffed;
  color: #52c41a;
}

.session-status.connecting {
  background-color: #fffbe6;
  color: #faad14;
}

.session-status.error,
.session-status.disconnected {
  background-color: #fff2f0;
  color: #ff4d4f;
}

.session-actions {
  display: flex;
  gap: 8px;
}

.btn-view, .btn-disconnect {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-view:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.btn-disconnect {
  color: #ff4d4f;
  border-color: #ffccc7;
}

.btn-disconnect:hover {
  background-color: #fff2f0;
}
</style>
