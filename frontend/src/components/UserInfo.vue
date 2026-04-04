<template>
  <div>
    <span class="username" @click="userinfoModelVisible = true" v-if="userInfo.username">{{ userInfo.username }}</span>
    <span class="username" @click="loginModelVisible = true" v-else>去登录</span>
  </div>

  <div v-if="loginModelVisible" class="modal-overlay" @click.self="loginModelVisible = false">
    <div class="modal">
      <h3>登录</h3>
      <div class="form">
        <div class="form-item">
          <label>用户名</label>
          <input type="text" v-model="loginForm.username" placeholder="请输入用户名">
        </div>
        <div class="form-item">
          <label>密码</label>
          <input type="password" v-model="loginForm.password" placeholder="请输入密码" @keyup.enter="handleLogin">
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-cancel" @click="loginModelVisible = false">取消</button>
        <button class="btn-confirm" @click="handleLogin" :disabled="btnLoading">
          {{ btnLoading ? '登录中...' : '登录' }}
        </button>
      </div>
      <p class="switch-text">
        没有账号？
        <a href="#" @click.prevent="toggleToRegister">注册</a>
      </p>
    </div>
  </div>

  <div v-if="registerModelVisible" class="modal-overlay" @click.self="registerModelVisible = false">
    <div class="modal">
      <h3>注册</h3>
      <div class="form">
        <div class="form-item">
          <label>用户名</label>
          <input type="text" v-model="registerForm.username" placeholder="请输入用户名">
        </div>
        <div class="form-item">
          <label>密码</label>
          <input type="password" v-model="registerForm.password" placeholder="请输入密码">
        </div>
        <div class="form-item">
          <label>确认密码</label>
          <input type="password" v-model="registerForm.confirmPassword" placeholder="请再次输入密码" @keyup.enter="handleRegister">
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-cancel" @click="registerModelVisible = false">取消</button>
        <button class="btn-confirm" @click="handleRegister" :disabled="btnLoading">
          {{ btnLoading ? '注册中...' : '注册' }}
        </button>
      </div>
      <p class="switch-text">
        已有账号？
        <a href="#" @click.prevent="toggleToLogin">登录</a>
      </p>
    </div>
  </div>

  <div v-if="userinfoModelVisible" class="modal-overlay" @click.self="userinfoModelVisible = false">
    <div class="modal">
      <h3>用户信息</h3>
      <div class="form">
        <div class="form-item">
          <label>用户名</label>
          <input type="text" :value="userInfo.username" disabled>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-cancel" @click="userinfoModelVisible = false">取消</button>
        <button class="btn-danger" @click="handleLogout" :disabled="btnLoading">
          {{ btnLoading ? '退出中...' : '退出登录' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>

import {onMounted, ref} from "vue";
import {getUserInfo, userLogin, userLogout, userRegister} from "../api/user.js";
import {message} from "ant-design-vue";

const userInfo = ref('')
const refreshUserInfo = () => {
  getUserInfo().then(res => {
    userInfo.value = res.data
  })
}
const btnLoading = ref(false);

onMounted(() => {
  refreshUserInfo()
})

const loginModelVisible = ref(false)
const loginForm = ref({
  username: '',
  password: ''
})
const handleLogin = () => {
  if (!loginForm.value.username) {
    message.error('请输入用户名')
    return
  }
  if (!loginForm.value.password) {
    message.error('请输入密码')
    return
  }
  btnLoading.value = true
  userLogin(loginForm.value).then(res => {
    if (res.code === 200) {
      refreshUserInfo()
      loginModelVisible.value = false
      loginForm.value = { username: '', password: '' }
      message.success('登录成功')
    }
  }).finally(() => {
    btnLoading.value = false
  })
}
const toggleToRegister = () => {
  loginModelVisible.value = false
  registerModelVisible.value = true
}

const registerModelVisible = ref(false)
const registerForm = ref({
  username: '',
  password: '',
  confirmPassword: ''
})
const handleRegister = () => {
  if (!registerForm.value.username) {
    message.error('请输入用户名')
    return
  }
  if (!registerForm.value.password) {
    message.error('请输入密码')
    return
  }
  if (!registerForm.value.confirmPassword) {
    message.error('请输入确认密码')
    return
  }
  if (registerForm.value.password !== registerForm.value.confirmPassword) {
    message.error('两次密码不一致')
    return
  }
  btnLoading.value = true
  userRegister(registerForm.value).then(res => {
    if (res.code === 200) {
      refreshUserInfo()
      registerModelVisible.value = false
      registerForm.value = { username: '', password: '', confirmPassword: '' }
      message.success('注册成功')
    }
  }).finally(() => {
    btnLoading.value = false
  })
}
const toggleToLogin = () => {
  loginModelVisible.value = true
  registerModelVisible.value = false
}

const userinfoModelVisible = ref(false)
const handleLogout = () => {
  btnLoading.value = true
  userLogout().then(res => {
    if (res.code === 200) {
      userInfo.value = ''
      userinfoModelVisible.value = false
      message.success('注销成功')
    }
  }).finally(() => {
    btnLoading.value = false
  })
}

</script>

<style scoped>
.username {
  color: #00d4ff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
}

.username:hover {
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
}

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
  margin-bottom: 24px;
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

.btn-danger {
  padding: 10px 24px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 0 15px rgba(239, 68, 68, 0.3);
}

.btn-danger:hover {
  box-shadow: 0 0 25px rgba(239, 68, 68, 0.5);
  transform: translateY(-2px);
}

.btn-danger:disabled {
  background: linear-gradient(135deg, #3d4571 0%, #2d3561 100%);
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

.switch-text {
  margin-top: 20px;
  text-align: center;
  color: #94a3b8;
  font-size: 14px;
}

.switch-text a {
  color: #00d4ff;
  text-decoration: none;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
  transition: all 0.3s ease;
}

.switch-text a:hover {
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
}
</style>
