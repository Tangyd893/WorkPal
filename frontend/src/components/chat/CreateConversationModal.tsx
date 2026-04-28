import { useEffect, useState } from 'react'
import type { CreateConversationDraft } from '../../types/chat'

interface CreateConversationModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (draft: CreateConversationDraft) => Promise<void>
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to create the conversation.'
}

function parseMemberIDs(value: string): number[] {
  const uniqueIDs = new Set<number>()

  value
    .split(',')
    .map((item) => Number.parseInt(item.trim(), 10))
    .filter((item) => Number.isInteger(item) && item > 0)
    .forEach((item) => uniqueIDs.add(item))

  return [...uniqueIDs]
}

export default function CreateConversationModal({ open, onClose, onSubmit }: CreateConversationModalProps) {
  const [mode, setMode] = useState<'private' | 'group'>('private')
  const [targetUserID, setTargetUserID] = useState('')
  const [groupName, setGroupName] = useState('')
  const [groupMembers, setGroupMembers] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setError('')
      return
    }

    setMode('private')
    setTargetUserID('')
    setGroupName('')
    setGroupMembers('')
    setError('')
    setSubmitting(false)
  }, [open])

  if (!open) {
    return null
  }

  const handleSubmit = async () => {
    setError('')

    if (mode === 'private') {
      const parsedUserID = Number.parseInt(targetUserID, 10)
      if (!Number.isInteger(parsedUserID) || parsedUserID <= 0) {
        setError('Please enter a valid user ID.')
        return
      }

      setSubmitting(true)
      try {
        await onSubmit({
          mode: 'private',
          targetUserId: parsedUserID,
        })
      } catch (submitError) {
        setError(getErrorMessage(submitError))
      } finally {
        setSubmitting(false)
      }
      return
    }

    const memberIDs = parseMemberIDs(groupMembers)
    if (memberIDs.length === 0) {
      setError('Please add at least one member ID.')
      return
    }

    setSubmitting(true)
    try {
      await onSubmit({
        mode: 'group',
        name: groupName.trim(),
        memberIds: memberIDs,
      })
    } catch (submitError) {
      setError(getErrorMessage(submitError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(15, 23, 42, 0.45)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 16,
        zIndex: 1000,
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 380,
          background: '#ffffff',
          borderRadius: 12,
          padding: 24,
          boxShadow: '0 24px 48px rgba(15, 23, 42, 0.18)',
        }}
      >
        <h3 style={{ margin: 0, fontSize: 20 }}>Create conversation</h3>
        <p style={{ margin: '8px 0 20px', color: '#6b7280', fontSize: 14 }}>
          Start a direct conversation or open a new group room.
        </p>

        <div style={{ display: 'flex', gap: 8, marginBottom: 20 }}>
          <button
            type="button"
            onClick={() => setMode('private')}
            style={{
              flex: 1,
              padding: '10px 12px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              background: mode === 'private' ? '#2563eb' : '#ffffff',
              color: mode === 'private' ? '#ffffff' : '#111827',
              fontWeight: 600,
            }}
          >
            Direct
          </button>
          <button
            type="button"
            onClick={() => setMode('group')}
            style={{
              flex: 1,
              padding: '10px 12px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              background: mode === 'group' ? '#2563eb' : '#ffffff',
              color: mode === 'group' ? '#ffffff' : '#111827',
              fontWeight: 600,
            }}
          >
            Group
          </button>
        </div>

        {mode === 'private' ? (
          <div className="form-item">
            <label htmlFor="target-user-id">Target user ID</label>
            <input
              id="target-user-id"
              type="number"
              value={targetUserID}
              onChange={(event) => setTargetUserID(event.target.value)}
              placeholder="Enter a numeric user ID"
            />
          </div>
        ) : (
          <>
            <div className="form-item">
              <label htmlFor="group-name">Group name</label>
              <input
                id="group-name"
                type="text"
                value={groupName}
                onChange={(event) => setGroupName(event.target.value)}
                placeholder="Optional group name"
              />
            </div>
            <div className="form-item">
              <label htmlFor="group-members">Member IDs</label>
              <input
                id="group-members"
                type="text"
                value={groupMembers}
                onChange={(event) => setGroupMembers(event.target.value)}
                placeholder="Example: 12,18,24"
              />
            </div>
          </>
        )}

        {error ? <div className="error-msg">{error}</div> : null}

        <div style={{ display: 'flex', gap: 8, marginTop: 20 }}>
          <button
            type="button"
            onClick={onClose}
            style={{
              flex: 1,
              padding: '10px 12px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              background: '#ffffff',
              color: '#111827',
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            disabled={submitting}
            style={{
              flex: 1,
              padding: '10px 12px',
              borderRadius: 8,
              background: '#2563eb',
              color: '#ffffff',
              opacity: submitting ? 0.7 : 1,
            }}
          >
            {submitting ? 'Creating...' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  )
}
