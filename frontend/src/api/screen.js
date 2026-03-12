// 屏幕流 API

// 获取会话的最新图像数据
export const getSessionImage = (sessionId) => {
    if (window.go?.main?.App?.GetSessionImage) {
        return window.go.main.App.GetSessionImage(sessionId)
    }
    return Promise.resolve({ code: 200, data: null })
}

// 开始接收屏幕流（返回一个可取消的函数）
export const startScreenStream = (sessionId, onFrame) => {
    // 使用轮询方式获取图像数据
    let isRunning = true
    let lastSequence = 0

    const poll = async () => {
        if (!isRunning) return

        try {
            const res = await getSessionImage(sessionId)
            if (res && res.code === 200 && res.data) {
                const data = res.data

                // 检查是否是新帧
                if (data.sequence > lastSequence || lastSequence === 0) {
                    lastSequence = data.sequence || 0

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
        if (isRunning) {
            setTimeout(poll, 33)
        }
    }

    // 开始轮询
    poll()

    // 返回停止函数
    return () => {
        isRunning = false
    }
}
