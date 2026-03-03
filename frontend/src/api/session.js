// 会话管理 API

// 获取所有会话
export const getAllSessions = () => {
    if (window.go?.main?.App?.GetAllSessions) {
        return window.go.main.App.GetAllSessions()
    }
    return Promise.resolve({ code: 200, data: [] })
}

// 获取单个会话
export const getSession = (sessionId) => {
    if (window.go?.main?.App?.GetSession) {
        return window.go.main.App.GetSession(sessionId)
    }
    return Promise.resolve({ code: 200, data: null })
}

// 根据设备码获取会话
export const getSessionByDeviceCode = (deviceCode) => {
    if (window.go?.main?.App?.GetSessionByDeviceCode) {
        return window.go.main.App.GetSessionByDeviceCode(deviceCode)
    }
    return Promise.resolve({ code: 200, data: null })
}

// 创建会话
export const createSession = (sessionId, deviceCode, deviceName, viewOnly) => {
    if (window.go?.main?.App?.CreateSession) {
        return window.go.main.App.CreateSession(sessionId, deviceCode, deviceName, viewOnly)
    }
    return Promise.resolve({ code: 200, data: { sessionId } })
}

// 移除会话
export const removeSession = (sessionId) => {
    if (window.go?.main?.App?.RemoveSession) {
        return window.go.main.App.RemoveSession(sessionId)
    }
    return Promise.resolve({ code: 200 })
}
