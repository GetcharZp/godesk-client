// 远程控制相关 API

// 发送控制请求
export const sendControlRequest = (params) => {
    // Wails 绑定的方法需要单独传参，不是对象
    if (window.go?.main?.App?.SendControlRequest) {
        return window.go.main.App.SendControlRequest(
            params.targetDeviceCode,
            params.targetPassword,
            params.requestControl
        )
    }
    // 模拟数据
    return Promise.resolve({
        code: 200,
        data: {
            accepted: true,
            sessionId: 'session_' + Date.now(),
            targetInfo: {
                name: params.targetDeviceCode.toString(),
                screenWidth: 1920,
                screenHeight: 1080
            }
        }
    })
}

// 断开控制连接
export const disconnectControl = (params) => {
    // Wails 绑定的方法需要单独传参
    if (window.go?.main?.App?.SendDisconnectNotify) {
        return window.go.main.App.SendDisconnectNotify(
            params.sessionId,
            params.targetDeviceCode
        )
    }
    return Promise.resolve({
        code: 200,
        data: { success: true }
    })
}

// 启动屏幕流
export const startScreenStream = (sessionId) => {
    // TODO: 调用后端方法
    return window.go?.main?.App?.StartScreenStream(sessionId) || Promise.resolve({
        code: 200
    })
}

// 停止屏幕流
export const stopScreenStream = (sessionId) => {
    // TODO: 调用后端方法
    return window.go?.main?.App?.StopScreenStream(sessionId) || Promise.resolve({
        code: 200
    })
}

// 发送鼠标移动事件
export const sendMouseMove = (sessionId, x, y) => {
    if (window.go?.main?.App?.SendMouseMove) {
        return window.go.main.App.SendMouseMove(sessionId, x, y)
    }
    return Promise.resolve({ code: 200 })
}

// 发送鼠标点击事件
export const sendMouseClick = (sessionId, x, y, button, action) => {
    if (window.go?.main?.App?.SendMouseClick) {
        return window.go.main.App.SendMouseClick(sessionId, x, y, button, action)
    }
    return Promise.resolve({ code: 200 })
}

// 发送鼠标滚轮事件
export const sendMouseScroll = (sessionId, x, y, deltaX, deltaY) => {
    if (window.go?.main?.App?.SendMouseScroll) {
        return window.go.main.App.SendMouseScroll(sessionId, x, y, deltaX, deltaY)
    }
    return Promise.resolve({ code: 200 })
}

// 发送键盘按下事件
export const sendKeyDown = (sessionId, key, modifiers = []) => {
    if (window.go?.main?.App?.SendKeyDown) {
        return window.go.main.App.SendKeyDown(sessionId, key, modifiers)
    }
    return Promise.resolve({ code: 200 })
}

// 发送键盘释放事件
export const sendKeyUp = (sessionId, key, modifiers = []) => {
    if (window.go?.main?.App?.SendKeyUp) {
        return window.go.main.App.SendKeyUp(sessionId, key, modifiers)
    }
    return Promise.resolve({ code: 200 })
}
