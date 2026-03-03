<template>
  <div class="remote-session-window">
    <div class="session-tabs" v-if="sessions.length > 0">
      <div
        v-for="session in sessions"
        :key="session.sessionId"
        class="session-tab"
        :class="{ active: currentSessionId === session.sessionId }"
        @click="switchSession(session.sessionId)"
      >
        <span class="tab-name">{{ session.deviceName }}</span>
        <span class="tab-status" :class="session.status"></span>
        <button class="tab-close" @click.stop="closeSession(session.sessionId)">×</button>
      </div>
    </div>

    <div class="session-content">
      <div v-if="currentSession" class="screen-container">
        <div class="toolbar">
          <div class="toolbar-left">
            <span class="device-name">{{ currentSession.deviceName }}</span>
            <span class="connection-status" :class="currentSession.status">
              {{ getStatusText(currentSession.status) }}
            </span>
          </div>
          <div class="toolbar-right">
            <button class="toolbar-btn" :class="{ active: currentSession.viewOnly }" @click="toggleViewOnly">
              <EyeOutlined /> 仅查看
            </button>
            <button class="toolbar-btn" @click="refreshScreen">
              <ReloadOutlined /> 刷新
            </button>
            <button class="toolbar-btn danger" @click="disconnectCurrent">
              <DisconnectOutlined /> 断开
            </button>
          </div>
        </div>

        <div class="screen-wrapper">
          <canvas
            ref="screenCanvas"
            class="screen-canvas"
            :width="currentSession.screenWidth || 1920"
            :height="currentSession.screenHeight || 1080"
          ></canvas>
          <div v-if="currentSession.status === 'connecting'" class="screen-overlay">
            <div class="loading-spinner"></div>
            <p>正在连接...</p>
          </div>
          <div v-if="currentSession.status === 'error'" class="screen-overlay error">
            <CloseCircleOutlined />
            <p>连接失败</p>
            <button class="retry-btn" @click="reconnect">重新连接</button>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        <DesktopOutlined class="empty-icon" />
        <p>暂无远程连接</p>
        <p class="empty-tip">请从主窗口发起远程控制</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  EyeOutlined,
  ReloadOutlined,
  DisconnectOutlined,
  DesktopOutlined,
  CloseCircleOutlined
} from '@ant-design/icons-vue'
import { disconnectControl } from '../api/channel.js'
import { getAllSessions, removeSession } from '../api/session.js'

const route = useRoute()

const sessions = ref([])
const currentSessionId = ref(null)
const screenCanvas = ref(null)

const currentSession = computed(() => {
  return sessions.value.find(s => s.sessionId === currentSessionId.value)
})

const getStatusText = (status) => {
  const statusMap = {
    'connecting': '连接中',
    'connected': '已连接',
    'error': '连接失败',
    'disconnected': '已断开'
  }
  return statusMap[status] || status
}

const loadSessions = async () => {
  try {
    const res = await getAllSessions()
    if (res && res.code === 200 && res.data) {
      sessions.value = res.data || []
      if (sessions.value.length > 0 && !currentSessionId.value) {
        currentSessionId.value = sessions.value[0].sessionId
      }
    }
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

const switchSession = (sessionId) => {
  currentSessionId.value = sessionId
}

const closeSession = async (sessionId) => {
  const session = sessions.value.find(s => s.sessionId === sessionId)
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

  await removeSession(sessionId)
  await loadSessions()

  if (currentSessionId.value === sessionId) {
    currentSessionId.value = sessions.value.length > 0 ? sessions.value[0].sessionId : null
  }
}

const toggleViewOnly = () => {
  if (currentSession.value) {
    currentSession.value.viewOnly = !currentSession.value.viewOnly
    message.info(currentSession.value.viewOnly ? '已切换到仅查看模式' : '已切换到控制模式')
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

let refreshTimer = null

onMounted(() => {
  const sessionId = route.params.sessionId
  if (sessionId) {
    currentSessionId.value = sessionId
  }

  loadSessions()
  refreshTimer = setInterval(loadSessions, 3000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

watch(() => currentSession.value?.lastImageData, (newData) => {
  if (newData && screenCanvas.value) {
    const ctx = screenCanvas.value.getContext('2d')
    const img = new Image()
    img.onload = () => {
      ctx.drawImage(img, 0, 0)
    }
    img.src = newData
  }
})
</script>

<style scoped>
.remote-session-window {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #1a1a1a;
  overflow: hidden;
}

.session-tabs {
  display: flex;
  background-color: #2d2d2d;
  padding: 0 8px;
  overflow-x: auto;
  flex-shrink: 0;
}

.session-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.3s;
  white-space: nowrap;
}

.session-tab:hover {
  background-color: #3d3d3d;
}

.session-tab.active {
  border-bottom-color: #1890ff;
  background-color: #1a1a1a;
}

.tab-name {
  font-size: 13px;
  color: #ccc;
}

.tab-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #666;
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
  width: 16px;
  height: 16px;
  border: none;
  background: none;
  cursor: pointer;
  color: #999;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  margin-left: 4px;
}

.tab-close:hover {
  color: #ff4d4f;
}

.session-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.screen-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.device-name {
  font-size: 14px;
  color: #fff;
  font-weight: 500;
}

.connection-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.connection-status.connected {
  background-color: #237804;
  color: #fff;
}

.connection-status.connecting {
  background-color: #ad6800;
  color: #fff;
}

.connection-status.error,
.connection-status.disconnected {
  background-color: #a8071a;
  color: #fff;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  padding: 6px 12px;
  background-color: #3d3d3d;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #ccc;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.3s;
}

.toolbar-btn:hover {
  background-color: #4d4d4d;
}

.toolbar-btn.active {
  background-color: #1890ff;
  color: white;
}

.toolbar-btn.danger {
  background-color: #ff4d4f;
  color: white;
}

.toolbar-btn.danger:hover {
  background-color: #ff7875;
}

.screen-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: auto;
  min-height: 0;
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
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.retry-btn {
  margin-top: 16px;
  padding: 8px 24px;
  background-color: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.retry-btn:hover {
  background-color: #40a9ff;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state p {
  margin: 0;
  font-size: 16px;
}

.empty-tip {
  margin-top: 8px !important;
  font-size: 14px !important;
  color: #555;
}
</style>
