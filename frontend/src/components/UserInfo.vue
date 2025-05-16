<template>
  <div>
    <span class="username" @click="userinfoModelVisible = true" v-if="userInfo.username" >{{ userInfo.username }}</span>
    <span class="username" @click="loginModelVisible = true" v-else>去登录</span>
  </div>

  <!-- 登录 -->
  <a-modal
      title="登录"
      v-model:open="loginModelVisible"
      :footer="null"
      width="450px"
  >
    <a-form
        :model="loginForm"
        @finish="handleLogin"
        :label-col="{ span: 5}"
        :rules="rules"
        style="margin-top: 20px"
    >
      <a-form-item
          label="用户名"
          name="username"
      >
        <a-input v-model:value="loginForm.username" placeholder="请输入用户名" />
      </a-form-item>
      <a-form-item
          label="密码"
          name="password"
      >
        <a-input-password v-model:value="loginForm.password" placeholder="请输入密码" />
      </a-form-item>
      <a-button type="primary" :loading="btnLoading" html-type="submit" block>登录</a-button>
    </a-form>
    <p style="margin-top: 16px; text-align: center;">
      没有账号？
      <a href="#" @click.prevent="toggleToRegister">注册</a>
    </p>
  </a-modal>

  <!-- 注册 -->
  <a-modal
      title="注册"
      v-model:open="registerModelVisible"
      :footer="null"
      width="450px"
  >
    <a-form
        :model="registerForm"
        @finish="handleRegister"
        :label-col="{ span: 5}"
        :rules="rules"
        style="margin-top: 20px"
    >
      <a-form-item
          label="用户名"
          name="username"
      >
        <a-input v-model:value="registerForm.username" placeholder="请输入用户名" />
      </a-form-item>
      <a-form-item
          label="密码"
          name="password"
      >
        <a-input-password v-model:value="registerForm.password" placeholder="请输入密码" />
      </a-form-item>
      <a-form-item
          label="确认密码"
          name="confirmPassword"
      >
        <a-input-password v-model:value="registerForm.confirmPassword" placeholder="请再次输入密码" />
      </a-form-item>
      <a-button type="primary" :loading="btnLoading" html-type="submit" block>注册</a-button>
    </a-form>

    <p style="margin-top: 16px; text-align: center;">
      已有账号？
      <a href="#" @click.prevent="toggleToLogin">登录</a>
    </p>
  </a-modal>

  <!-- 用户信息 -->
  <a-modal
      title="用户信息"
      v-model:open="userinfoModelVisible"
      @finish="handleLogout"
      :footer="null"
      width="450px"
  >
    <a-form
        :label-col="{ span: 5}"
        style="margin-top: 20px"
    >
      <a-form-item
          label="用户名"
          name="username"
      >
        <a-input v-model:value="userInfo.username" disabled placeholder="请输入用户名" />
      </a-form-item>
      <a-button type="primary" block @click="handleLogout">退出登录</a-button>
    </a-form>
  </a-modal>
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

const rules = {
  username: [
    {
      required: true,
      message: '请输入用户名',
      trigger: 'blur'
    }
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: 'blur'
    }
  ],
  confirmPassword:  [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value) {
          callback(new Error('请输入确认密码'))
        } else if (value !== registerForm.value.password) {
          callback(new Error('两次密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const loginModelVisible = ref(false)
const loginForm = ref({
  username: '',
  password: ''
})
const handleLogin = () => {
  btnLoading.value = true
  userLogin(loginForm.value).then(res => {
    if (res.code === 200) {
      refreshUserInfo()
      loginModelVisible.value = false
      btnLoading.value = false
      message.success('登录成功')
    }
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
  btnLoading.value = true
  userRegister(registerForm.value).then(res => {
    if (res.code === 200) {
      refreshUserInfo()
      registerModelVisible.value = false
      btnLoading.value = false
      message.success('注册成功')
    }
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
      refreshUserInfo()
      userinfoModelVisible.value = false
      btnLoading.value = false
      message.success('注销成功')
    }
  })
}

</script>

<style scoped>
.username {
  color: #409EFF;
  font-size: 14px;
  cursor: pointer;
}
</style>