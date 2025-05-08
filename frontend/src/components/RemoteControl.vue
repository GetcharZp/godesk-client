<template>
  <div class="remote-control">
    <h2 class="page-title">远程控制</h2>

    <div class="control-panel">
      <div class="control-section">
        <label class="control-option">
          允许控制本设备
        </label>

        <div class="device-code">
          <p>本设备识别码</p>
          <p class="code">{{deviceInfo.code}}</p>
        </div>
      </div>

      <div class="auth-section">
        <p>验证码</p>
        <div class="auth-codes">
          <span class="code">{{ deviceInfo.password }}</span>
        </div>
      </div>

      <div class="remote-operate">
        <h3>远程控制设备</h3>
        <div class="input-group">
          <input type="text" placeholder="请输入伙伴识别码">
          <input type="text" placeholder="验证码">
        </div>
        <div class="action-buttons">
          <button class="btn remote-desktop">远程桌面</button>
          <button class="btn remote-files">远程文件</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref} from "vue";
import {getDeviceInfo} from "../api/device.js"

const deviceInfo = ref('')

onMounted(async () => {
  getDeviceInfo().then(res => {
    if (res) {
      deviceInfo.value = res.data
    }
  })
})

</script>

<style scoped>
.remote-control {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  color: #333;
  margin-bottom: 20px;
  font-size: 20px;
}

.control-panel {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.control-section {
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.control-option {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  font-size: 16px;
}

.device-code p:first-child {
  color: #666;
  font-size: 14px;
  margin-bottom: 5px;
}

.device-code .code {
  font-size: 22px;
  font-weight: 600;
  color: #333;
}

.auth-section p:first-child {
  color: #666;
  font-size: 14px;
  margin-bottom: 10px;
}

.auth-codes {
  display: flex;
  gap: 15px;
}

.auth-codes .code {
  font-family: monospace;
  font-size: 16px;
  color: #333;
}

.remote-operate {
  margin-top: 30px;
}

.remote-operate h3 {
  font-size: 16px;
  margin-bottom: 15px;
  color: #333;
}

.input-group {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.input-group input {
  flex: 1;
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.input-group input::placeholder {
  color: #999;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.remote-desktop {
  background-color: #409EFF;
  color: white;
}

.remote-files {
  background-color: #f0f0f0;
  color: #333;
}
</style>