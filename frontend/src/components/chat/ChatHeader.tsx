interface ChatHeaderProps {
  connected: boolean
  username: string | null
  onLogout: () => void
}

export default function ChatHeader({ connected, username, onLogout }: ChatHeaderProps) {
  return (
    <header
      style={{
        height: 64,
        padding: '0 20px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        background: '#ffffff',
        borderBottom: '1px solid #e5e7eb',
      }}
    >
      <div>
        <div style={{ fontSize: 18, fontWeight: 700 }}>WorkPal</div>
        <div style={{ fontSize: 12, color: '#6b7280' }}>Team messaging workspace</div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span
          style={{
            padding: '4px 10px',
            borderRadius: 999,
            fontSize: 12,
            fontWeight: 600,
            background: connected ? '#dcfce7' : '#fee2e2',
            color: connected ? '#166534' : '#991b1b',
          }}
        >
          {connected ? 'Connected' : 'Disconnected'}
        </span>
        <span style={{ color: '#374151', fontSize: 14 }}>{username || 'Unknown user'}</span>
        <button
          type="button"
          onClick={onLogout}
          style={{
            padding: '8px 12px',
            borderRadius: 8,
            border: '1px solid #d1d5db',
            background: '#ffffff',
            color: '#111827',
          }}
        >
          Sign out
        </button>
      </div>
    </header>
  )
}
