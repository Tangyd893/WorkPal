import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore, useConvStore, useWSStore, ChatMessage, Conversation } from '../hooks/useAuthStore'
import request from '../api/request'

// WebSocket 消息格式
interface WSMessage {
  type: string
  from?: number
  to?: string
  conv_id?: number
  content?: string
  seq?: number
  created_at?: string
  user_id?: number
  status?: string
}

export default function ChatPage() {
  const navigate = useNavigate()
  const { username, userId, token, logout } = useAuthStore()
  const { conversations, setConversations, activeConvID, setActiveConvID } = useConvStore()
  const { connected, setConnected, addMessage, messages, setMessages } = useWSStore()

  const [input, setInput] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [createType, setCreateType] = useState<'private' | 'group'>('private')
  const [targetUID, setTargetUID] = useState('')
  const [groupName, setGroupName] = useState('')
  const [groupMembers, setGroupMembers] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  // 加载会话列表
  const loadConversations = async () => {
    try {
      const res = await request.get<any, any>('/conversations')
      setConversations(res.data || [])
    } catch (err) {
      console.error('加载会话列表失败', err)
    }
  }

  // 连接 WebSocket
  const connectWS = () => {
    if (!token) return
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const ws = new WebSocket(`${protocol}//${host}/ws?token=${token}`)

    ws.onopen = () => {
      console.log('[WS] 已连接')
      setConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        if (msg.type === 'chat' && msg.conv_id) {
          const chatMsg: ChatMessage = {
            id: Date.now(),
            conv_id: msg.conv_id!,
            sender_id: msg.from!,
            type: 1,
            content: typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content),
            created_at: msg.created_at || new Date().toISOString(),
          }
          addMessage(msg.conv_id, chatMsg)
        }
      } catch (e) {
        console.error('解析 WS 消息失败', e)
      }
    }

    ws.onclose = () => {
      console.log('[WS] 已断开')
      setConnected(false)
      // 30秒后重连
      setTimeout(connectWS, 30000)
    }

    ws.onerror = (err) => {
      console.error('[WS] 错误', err)
    }

    wsRef.current = ws
  }

  useEffect(() => {
    loadConversations()
    connectWS()
    return () => {
      wsRef.current?.close()
    }
  }, [])

  // 滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, activeConvID])

  // 加载历史消息
  const loadMessages = async (convID: number) => {
    try {
      const res = await request.get<any, any>(`/conversations/${convID}/messages`)
      setMessages(convID, res.data || [])
    } catch (err) {
      console.error('加载消息失败', err)
    }
  }

  const handleSelectConv = (conv: Conversation) => {
    setActiveConvID(conv.id)
    if (!messages[conv.id]) {
      loadMessages(conv.id)
    }
  }

  // 发送消息
  const handleSend = async () => {
    if (!input.trim() || !activeConvID) return
    try {
      const res = await request.post<any, any>(`/conversations/${activeConvID}/messages`, {
        type: 1,
        content: input.trim(),
      })
      const msg: ChatMessage = res.data
      addMessage(activeConvID, msg)
      setInput('')
    } catch (err) {
      console.error('发送失败', err)
    }
  }

  // 创建会话
  const handleCreateConv = async () => {
    try {
      if (createType === 'private') {
        const res = await request.post<any, any>('/conversations', {
          type: 1,
          target_uid: parseInt(targetUID),
        })
        await loadConversations()
        if (res.data?.id) {
          setActiveConvID(res.data.id)
          setMessages(res.data.id, [])
        }
      } else {
        const memberIDs = groupMembers.split(',').map(s => parseInt(s.trim())).filter(n => n > 0)
        const res = await request.post<any, any>('/conversations', {
          type: 2,
          name: groupName || '群聊',
          member_ids: memberIDs,
        })
        await loadConversations()
        if (res.data?.id) {
          setActiveConvID(res.data.id)
          setMessages(res.data.id, [])
        }
      }
      setShowCreate(false)
      setTargetUID('')
      setGroupName('')
      setGroupMembers('')
    } catch (err) {
      console.error('创建会话失败', err)
    }
  }

  const currentMessages = activeConvID ? (messages[activeConvID] || []) : []
  const currentConv = conversations.find(c => c.id === activeConvID)

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column', background: '#f5f5f5' }}>
      {/* 顶部栏 */}
      <header style={{
        height: 56, background: 'white', borderBottom: '1px solid #e5e7eb',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '0 20px'
      }}>
        <span style={{ fontSize: 16, fontWeight: 600 }}>WorkPal</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{
            fontSize: 12, padding: '2px 8px', borderRadius: 10,
            background: connected ? '#d1fae5' : '#fee2e2',
            color: connected ? '#065f46' : '#991b1b'
          }}>
            {connected ? '在线' : '离线'}
          </span>
          <span style={{ color: '#666' }}>{username}</span>
          <button
            onClick={handleLogout}
            style={{ padding: '4px 12px', borderRadius: 6, background: '#f5f5f5', fontSize: 13, border: 'none', cursor: 'pointer' }}
          >
            退出
          </button>
        </div>
      </header>

      {/* 主体 */}
      <div style={{ flex: 1, display: 'flex' }}>
        {/* 侧边栏 */}
        <aside style={{ width: 260, background: 'white', borderRight: '1px solid #e5e7eb', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: 12, borderBottom: '1px solid #f0f0f0' }}>
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                type="text"
                placeholder="搜索会话"
                style={{
                  flex: 1, padding: '8px 12px', borderRadius: 6,
                  border: '1px solid #e5e7eb', fontSize: 13
                }}
              />
              <button
                onClick={() => setShowCreate(true)}
                style={{ padding: '6px 12px', borderRadius: 6, background: '#2563eb', color: 'white', fontSize: 12, border: 'none', cursor: 'pointer' }}
              >
                新建
              </button>
            </div>
          </div>

          {/* 会话列表 */}
          <div style={{ flex: 1, overflowY: 'auto' }}>
            {conversations.length === 0 ? (
              <div style={{ padding: 24, textAlign: 'center', color: '#999', fontSize: 13 }}>
                暂无会话
              </div>
            ) : (
              conversations.map(conv => (
                <div
                  key={conv.id}
                  onClick={() => handleSelectConv(conv)}
                  style={{
                    padding: '12px 16px', cursor: 'pointer',
                    background: conv.id === activeConvID ? '#eff6ff' : 'transparent',
                    borderBottom: '1px solid #f9fafb'
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <div style={{
                      width: 40, height: 40, borderRadius: '50%', background: '#e5e7eb',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: 14, color: '#666'
                    }}>
                      {conv.type === 1 ? '💬' : '👥'}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 2 }}>
                        {conv.type === 1 ? `用户${conv.owner_id}` : (conv.name || '群聊')}
                      </div>
                      <div style={{ fontSize: 12, color: '#999', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {conv.type === 2 ? `${conv.name}` : '私聊'}
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </aside>

        {/* 聊天主区 */}
        <main style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          {activeConvID && currentConv ? (
            <>
              {/* 聊天头部 */}
              <div style={{ padding: '12px 20px', background: 'white', borderBottom: '1px solid #e5e7eb', display: 'flex', alignItems: 'center' }}>
                <span style={{ fontWeight: 600 }}>
                  {currentConv.type === 1 ? `与用户${currentConv.owner_id}的私聊` : (currentConv.name || '群聊')}
                </span>
              </div>

              {/* 消息区 */}
              <div style={{ flex: 1, overflowY: 'auto', padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
                {currentMessages.map((msg, idx) => (
                  <div
                    key={msg.id || idx}
                    style={{
                      display: 'flex',
                      justifyContent: msg.sender_id === userId ? 'flex-end' : 'flex-start'
                    }}
                  >
                    <div style={{
                      maxWidth: '70%',
                      padding: '10px 14px',
                      borderRadius: 12,
                      background: msg.sender_id === userId ? '#2563eb' : 'white',
                      color: msg.sender_id === userId ? 'white' : '#333',
                      boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                      wordBreak: 'break-word'
                    }}>
                      <div>{msg.content}</div>
                      <div style={{ fontSize: 11, opacity: 0.7, marginTop: 4, textAlign: 'right' }}>
                        {new Date(msg.created_at).toLocaleTimeString()}
                      </div>
                    </div>
                  </div>
                ))}
                <div ref={messagesEndRef} />
              </div>

              {/* 输入框 */}
              <div style={{ padding: 16, background: 'white', borderTop: '1px solid #e5e7eb' }}>
                <div style={{ display: 'flex', gap: 8 }}>
                  <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSend()}
                    placeholder="输入消息..."
                    style={{
                      flex: 1, padding: '10px 14px', borderRadius: 8,
                      border: '1px solid #e5e7eb', fontSize: 14
                    }}
                  />
                  <button
                    onClick={handleSend}
                    style={{
                      padding: '10px 20px', borderRadius: 8, background: '#2563eb',
                      color: 'white', fontSize: 14, border: 'none', cursor: 'pointer'
                    }}
                  >
                    发送
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
              选择一个会话开始聊天
            </div>
          )}
        </main>
      </div>

      {/* 新建会话弹窗 */}
      {showCreate && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000
        }}>
          <div style={{ background: 'white', borderRadius: 12, padding: 24, width: 360 }}>
            <h3 style={{ marginTop: 0 }}>新建会话</h3>
            <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
              <button
                onClick={() => setCreateType('private')}
                style={{
                  flex: 1, padding: 8, borderRadius: 6, border: 'none', cursor: 'pointer',
                  background: createType === 'private' ? '#2563eb' : '#f5f5f5', color: createType === 'private' ? 'white' : '#333'
                }}
              >
                私聊
              </button>
              <button
                onClick={() => setCreateType('group')}
                style={{
                  flex: 1, padding: 8, borderRadius: 6, border: 'none', cursor: 'pointer',
                  background: createType === 'group' ? '#2563eb' : '#f5f5f5', color: createType === 'group' ? 'white' : '#333'
                }}
              >
                群聊
              </button>
            </div>

            {createType === 'private' ? (
              <div className="form-item">
                <label>目标用户ID</label>
                <input
                  type="number"
                  value={targetUID}
                  onChange={(e) => setTargetUID(e.target.value)}
                  placeholder="输入用户ID"
                  style={{ width: '100%', padding: '8px 12px', borderRadius: 6, border: '1px solid #e5e7eb', fontSize: 14 }}
                />
              </div>
            ) : (
              <>
                <div className="form-item">
                  <label>群名</label>
                  <input
                    type="text"
                    value={groupName}
                    onChange={(e) => setGroupName(e.target.value)}
                    placeholder="输入群名（选填）"
                    style={{ width: '100%', padding: '8px 12px', borderRadius: 6, border: '1px solid #e5e7eb', fontSize: 14 }}
                  />
                </div>
                <div className="form-item" style={{ marginTop: 12 }}>
                  <label>成员ID（逗号分隔）</label>
                  <input
                    type="text"
                    value={groupMembers}
                    onChange={(e) => setGroupMembers(e.target.value)}
                    placeholder="1,2,3"
                    style={{ width: '100%', padding: '8px 12px', borderRadius: 6, border: '1px solid #e5e7eb', fontSize: 14 }}
                  />
                </div>
              </>
            )}

            <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
              <button
                onClick={() => setShowCreate(false)}
                style={{ flex: 1, padding: 10, borderRadius: 6, background: '#f5f5f5', border: 'none', cursor: 'pointer' }}
              >
                取消
              </button>
              <button
                onClick={handleCreateConv}
                style={{ flex: 1, padding: 10, borderRadius: 6, background: '#2563eb', color: 'white', border: 'none', cursor: 'pointer' }}
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
