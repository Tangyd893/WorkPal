import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../hooks/useAuthStore'

export default function ChatPage() {
  const navigate = useNavigate()
  const { username, logout } = useAuthStore()
  const [message, setMessage] = useState('')

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 顶部栏 */}
      <header style={{
        height: 56, background: 'white', borderBottom: '1px solid #e5e7eb',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '0 20px'
      }}>
        <span style={{ fontSize: 16, fontWeight: 600 }}>WorkPal</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ color: '#666' }}>{username}</span>
          <button
            onClick={handleLogout}
            style={{ padding: '4px 12px', borderRadius: 6, background: '#f5f5f5', fontSize: 13 }}
          >
            退出
          </button>
        </div>
      </header>

      {/* 聊天区域 */}
      <div style={{ flex: 1, display: 'flex' }}>
        {/* 侧边栏 */}
        <aside style={{ width: 240, background: 'white', borderRight: '1px solid #e5e7eb' }}>
          <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
            <input
              type="text"
              placeholder="搜索"
              style={{
                width: '100%', padding: '8px 12px', borderRadius: 6,
                border: '1px solid #e5e7eb', fontSize: 13
              }}
            />
          </div>
          {/* TODO: 会话列表 */}
          <div style={{ padding: 16, color: '#999', fontSize: 13 }}>
            暂无会话
          </div>
        </aside>

        {/* 聊天主区 */}
        <main style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
            选择一个会话开始聊天
          </div>
          {/* 输入框 */}
          <div style={{ padding: 16, borderTop: '1px solid #e5e7eb', background: 'white' }}>
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                type="text"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="输入消息..."
                style={{
                  flex: 1, padding: '10px 14px', borderRadius: 8,
                  border: '1px solid #e5e7eb', fontSize: 14
                }}
              />
              <button className="btn-primary" style={{ width: 80, padding: '10px' }}>
                发送
              </button>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
