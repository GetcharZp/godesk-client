// 屏幕流 API

export const getSessionImage = (sessionId) => {
    if (window.go?.main?.App?.GetSessionImage) {
        return window.go.main.App.GetSessionImage(sessionId)
    }
    return Promise.resolve({ code: 200, data: null })
}

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
        if (existingState.currentObjectUrl) {
            URL.revokeObjectURL(existingState.currentObjectUrl)
        }
        streamStates.delete(sessionId)
    }

    const state = {
        isRunning: true,
        lastSequence: 0,
        lastTimestamp: 0,
        pollTimer: null,
        currentInterval: MIN_POLL_INTERVAL,
        staticCount: 0,
        lastFrameTime: Date.now(),
        currentObjectUrl: null
    }
    streamStates.set(sessionId, state)

    const revokeCurrentObjectUrl = () => {
        if (state.currentObjectUrl) {
            URL.revokeObjectURL(state.currentObjectUrl)
            state.currentObjectUrl = null
        }
    }

    const loadSessionImageUrl = async (imageUrl) => {
        if (!imageUrl) {
            throw new Error('image url is empty')
        }

        const response = await fetch(imageUrl, {
            cache: 'no-store'
        })
        if (!response.ok) {
            throw new Error(`image request failed: ${response.status}`)
        }

        const contentType = response.headers.get('content-type') || ''
        if (!contentType.includes('image/jpeg')) {
            throw new Error(`unexpected content-type: ${contentType || 'unknown'}`)
        }

        const imageBlob = await response.blob()
        revokeCurrentObjectUrl()
        state.currentObjectUrl = URL.createObjectURL(imageBlob)
        return state.currentObjectUrl
    }

    const poll = async () => {
        if (!state.isRunning) return

        try {
            const res = await getSessionImage(sessionId)
            if (res && res.code === 200 && res.data) {
                const data = res.data
                const nextSequence = data.sequence || 0
                const nextTimestamp = data.timestamp || 0
                const isSequenceReset =
                    state.lastSequence > 0 &&
                    nextSequence > 0 &&
                    nextSequence < state.lastSequence &&
                    nextTimestamp > state.lastTimestamp

                if (nextSequence > state.lastSequence || state.lastSequence === 0 || isSequenceReset) {
                    if (isSequenceReset) {
                        console.info('Screen stream sequence reset detected, restarting frame acceptance', {
                            sessionId,
                            previousSequence: state.lastSequence,
                            nextSequence,
                            previousTimestamp: state.lastTimestamp,
                            nextTimestamp
                        })
                    }

                    state.lastSequence = nextSequence
                    state.lastTimestamp = nextTimestamp
                    state.staticCount = 0
                    state.currentInterval = MIN_POLL_INTERVAL
                    state.lastFrameTime = Date.now()

                    if (onFrame && data.codec === 'jpeg' && data.hasImage) {
                        const imageUrl = await loadSessionImageUrl(data.imageUrl)
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
        revokeCurrentObjectUrl()
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
        if (state.currentObjectUrl) {
            URL.revokeObjectURL(state.currentObjectUrl)
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
