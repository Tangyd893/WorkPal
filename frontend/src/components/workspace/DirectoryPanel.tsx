import type { AppTranslations } from '../../i18n'
import type { TeamProfileMeta, WorkspaceUser } from '../../types/workspace'

interface DirectoryPanelProps {
  users: WorkspaceUser[]
  query: string
  currentUserId: number | null
  text: AppTranslations
  onQueryChange: (query: string) => void
  getProfileMeta: (username: string) => TeamProfileMeta
}

export default function DirectoryPanel({
  users,
  query,
  currentUserId,
  text,
  onQueryChange,
  getProfileMeta,
}: DirectoryPanelProps) {
  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.directory.title}</h2>
          <p>{text.directory.subtitle}</p>
        </div>
        <div className="toolbar-search">
          <input
            type="search"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder={text.directory.searchPlaceholder}
          />
        </div>
      </div>

      {users.length === 0 ? (
        <div className="empty-panel">{text.directory.noResults}</div>
      ) : (
        <div className="directory-grid">
          {users.map((user) => {
            const meta = getProfileMeta(user.username)
            const current = user.id === currentUserId

            return (
              <article key={user.id} className={current ? 'person-card highlight' : 'person-card'}>
                <div className="person-card-header">
                  <div>
                    <strong>{user.nickname || user.username}</strong>
                    <p>@{user.username}</p>
                  </div>
                  {current ? <span className="chip">{text.directory.currentUser}</span> : null}
                </div>
                <dl className="meta-pairs">
                  <div>
                    <dt>{text.directory.idLabel}</dt>
                    <dd>{user.id}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.emailLabel}</dt>
                    <dd>{user.email || text.common.unavailable}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.phoneLabel}</dt>
                    <dd>{user.phone || text.common.unavailable}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.roleLabel}</dt>
                    <dd>{meta.role}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.departmentLabel}</dt>
                    <dd>{meta.department}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.locationLabel}</dt>
                    <dd>{meta.location}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.focusLabel}</dt>
                    <dd>{meta.focus}</dd>
                  </div>
                </dl>
              </article>
            )
          })}
        </div>
      )}
    </section>
  )
}
