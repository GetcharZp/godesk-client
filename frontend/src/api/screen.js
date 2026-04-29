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

const MIN_POLL_INTERVAL = 33
const MAX_POLL_INTERVAL = 500
const STATIC_THRESHOLD = 5

export const startScreenStream = (sessionId, onFrame) => {
    if (streamStates.has(sessionId)) {
        const existingState = streamStates.get(sessionId)
        existingState.isRunning = false
        if (existingState.pollTimer) {
            clearTimeout(existingState.pollTimer)
        }
        streamStates.delete(sessionId)
    }

    const state = {
        isRunning: true,
        lastSequence: 0,
        pollTimer: null,
        currentInterval: MIN_POLL_INTERVAL,
        staticCount: 0,
        lastFrameTime: Date.now()
    }
    streamStates.set(sessionId, state)

    const poll = async () => {
        if (!state.isRunning) return

        try {
            const res = await getSessionImage(sessionId)
            if (res && res.code === 200 && res.data) {
                const data = res.data

                if (data.sequence > state.lastSequence || state.lastSequence === 0) {
                    state.lastSequence = data.sequence || 0
                    state.staticCount = 0
                    state.currentInterval = MIN_POLL_INTERVAL
                    state.lastFrameTime = Date.now()

                    if (onFrame && data.imageData) {
                        const imageUrl = 'data:image/jpeg;base64,' + data.imageData
                        onFrame(imageUrl, data)
                    }
                } else {
                    state.staticCount++
                    if (state.staticCount > STATIC_THRESHOLD * 3) {
                        state.currentInterval = MAX_POLL_INTERVAL
                    } else if (state.staticCount > STATIC_THRESHOLD) {
                        state.currentInterval = 200
                    } else {
                        state.currentInterval = Math.min(
                            state.currentInterval * 1.5,
                            MAX_POLL_INTERVAL
                        )
                    }
                }
            }
        } catch (error) {
            console.error('获取屏幕数据失败:', error)
        }

        if (state.isRunning) {
            state.pollTimer = setTimeout(poll, state.currentInterval)
        }
    }

    poll()

    return () => {
        state.isRunning = false
        if (state.pollTimer) {
            clearTimeout(state.pollTimer)
        }
        streamStates.delete(sessionId)
    }
}

export const stopScreenStream = (sessionId) => {
    const state = streamStates.get(sessionId)
    if (state) {
        state.isRunning = false
        if (state.pollTimer) {
            clearTimeout(state.pollTimer)
        }
        streamStates.delete(sessionId)
    }
}

export const getStreamStats = (sessionId) => {
    const state = streamStates.get(sessionId)
    if (state) {
        return {
            isRunning: state.isRunning,
            lastSequence: state.lastSequence,
            currentInterval: state.currentInterval,
            staticCount: state.staticCount
        }
    }
    return null
}
