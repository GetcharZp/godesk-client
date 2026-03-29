// 屏幕流 API

// 获取会话的最新图像数据
export const getSessionImage = (sessionId) => {
    if (window.go?.main?.App?.GetSessionImage) {
        return window.go.main.App.GetSessionImage(sessionId)
    }
    return Promise.resolve({ code: 200, data: null })
}

// 存储每个会话的轮询状态
const streamStates = new Map()

// 开始接收屏幕流（返回一个可取消的函数）
export const startScreenStream = (sessionId, onFrame) => {
    // 如果已有该会话的轮询，先停止
    if (streamStates.has(sessionId)) {
        const existingState = streamStates.get(sessionId)
        existingState.isRunning = false
        streamStates.delete(sessionId)
    }

    const state = {
        isRunning: true,
        lastSequence: 0,
        pollTimer: null
    }
    streamStates.set(sessionId, state)

    const poll = async () => {
        if (!state.isRunning) return

        try {
            const res = await getSessionImage(sessionId)
            if (res && res.code === 200 && res.data) {
                const data = res.data

                // 检查是否是新帧
                if (data.sequence > state.lastSequence || state.lastSequence === 0) {
                    state.lastSequence = data.sequence || 0

                    if (onFrame && data.imageData) {
                        const imageUrl = 'data:image/jpeg;base64,' + data.imageData
                        onFrame(imageUrl, data)
                    }
                }
            }
        } catch (error) {
            console.error('获取屏幕数据失败:', error)
        }

        // 继续轮询（约 30fps）
        if (state.isRunning) {
            state.pollTimer = setTimeout(poll, 33)
        }
    }

    // 开始轮询
    poll()

    // 返回停止函数
    return () => {
        state.isRunning = false
        if (state.pollTimer) {
            clearTimeout(state.pollTimer)
        }
        streamStates.delete(sessionId)
    }
}
