import { useEffect, useMemo, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import type { CreateConversationDraft } from '../../types/chat'
import type { WorkspaceUser } from '../../types/workspace'

interface CreateConversationModalProps {
  open: boolean
  users: WorkspaceUser[]
  currentUserId: number | null
  labels: AppTranslations['chat']
  common: AppTranslations['common']
  onClose: () => void
  onSubmit: (draft: CreateConversationDraft) => Promise<void>
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to create the conversation.'
}

export default function CreateConversationModal({
  open,
  users,
  currentUserId,
  labels,
  common,
  onClose,
  onSubmit,
}: CreateConversationModalProps) {
  const teammates = useMemo(
    () => users.filter((user) => user.id !== currentUserId).sort((left, right) => left.username.localeCompare(right.username)),
    [currentUserId, users],
  )

  const [mode, setMode] = useState<'private' | 'group'>('private')
  const [selectedDirectUserId, setSelectedDirectUserId] = useState<number | null>(null)
  const [groupName, setGroupName] = useState('')
  const [groupMemberIds, setGroupMemberIds] = useState<number[]>([])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) {
      setMode('private')
      setSelectedDirectUserId(null)
      setGroupName('')
      setGroupMemberIds([])
      setError('')
      setSubmitting(false)
      return
    }

    setError('')
    setSelectedDirectUserId((current) => current ?? teammates[0]?.id ?? null)
  }, [open, teammates])

  if (!open) {
    return null
  }

  const toggleGroupMember = (userId: number) => {
    setGroupMemberIds((current) =>
      current.includes(userId) ? current.filter((item) => item !== userId) : [...current, userId].sort((left, right) => left - right),
    )
  }

  const handleSubmit = async () => {
    setError('')

    if (mode === 'private') {
      if (!selectedDirectUserId) {
        setError(labels.invalidDirect)
        return
      }

      setSubmitting(true)
      try {
        await onSubmit({
          mode: 'private',
          targetUserId: selectedDirectUserId,
        })
      } catch (submitError) {
        setError(getErrorMessage(submitError))
      } finally {
        setSubmitting(false)
      }
      return
    }

    if (groupMemberIds.length === 0) {
      setError(labels.invalidGroup)
      return
    }

    setSubmitting(true)
    try {
      await onSubmit({
        mode: 'group',
        name: groupName.trim() || labels.groupChat,
        memberIds: groupMemberIds,
      })
    } catch (submitError) {
      setError(getErrorMessage(submitError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="dialog-scrim">
      <div className="dialog-panel" role="dialog" aria-modal="true" aria-labelledby="create-conversation-title">
        <div className="dialog-header">
          <div>
            <h3 id="create-conversation-title">{labels.createTitle}</h3>
            <p>{labels.createSubtitle}</p>
          </div>
        </div>

        <div className="segmented-control dialog-segment">
          <button
            type="button"
            className={mode === 'private' ? 'segment-button active' : 'segment-button'}
            aria-pressed={mode === 'private'}
            onClick={() => setMode('private')}
          >
            {labels.direct}
          </button>
          <button
            type="button"
            className={mode === 'group' ? 'segment-button active' : 'segment-button'}
            aria-pressed={mode === 'group'}
            onClick={() => setMode('group')}
          >
            {labels.group}
          </button>
        </div>

        {teammates.length === 0 ? (
          <div className="empty-panel">{labels.noTeamMembers}</div>
        ) : mode === 'private' ? (
          <div className="form-stack">
            <div className="form-copy">
              <strong>{labels.directTarget}</strong>
              <span>{labels.directTargetHint}</span>
            </div>
            <div className="choice-list">
              {teammates.map((user) => {
                const selected = user.id === selectedDirectUserId

                return (
                  <button
                    key={user.id}
                    type="button"
                    className={selected ? 'choice-card selected' : 'choice-card'}
                    aria-pressed={selected}
                    onClick={() => setSelectedDirectUserId(user.id)}
                  >
                    <div>
                      <strong>{user.nickname || user.username}</strong>
                      <span>@{user.username}</span>
                    </div>
                    <span className="choice-meta">#{user.id}</span>
                  </button>
                )
              })}
            </div>
          </div>
        ) : (
          <div className="form-stack">
            <div className="form-item">
              <label htmlFor="group-name">{labels.groupName}</label>
              <input
                id="group-name"
                type="text"
                value={groupName}
                onChange={(event) => setGroupName(event.target.value)}
                placeholder={labels.groupNamePlaceholder}
              />
            </div>
            <div className="form-copy">
              <strong>{labels.groupMembers}</strong>
              <span>{labels.groupMembersHint}</span>
            </div>
            <div className="choice-list">
              {teammates.map((user) => {
                const selected = groupMemberIds.includes(user.id)

                return (
                  <button
                    key={user.id}
                    type="button"
                    className={selected ? 'choice-card selected' : 'choice-card'}
                    aria-pressed={selected}
                    onClick={() => toggleGroupMember(user.id)}
                  >
                    <div>
                      <strong>{user.nickname || user.username}</strong>
                      <span>@{user.username}</span>
                    </div>
                    <span className="choice-meta">#{user.id}</span>
                  </button>
                )
              })}
            </div>
          </div>
        )}

        {error ? <div className="error-msg" role="alert">{error}</div> : null}

        <div className="dialog-actions">
          <button type="button" className="secondary-button" onClick={onClose}>
            {common.cancel}
          </button>
          <button type="button" className="primary-button" onClick={handleSubmit} disabled={submitting || teammates.length === 0}>
            {submitting ? labels.creating : labels.createAction}
          </button>
        </div>
      </div>
    </div>
  )
}
