import type { AppTranslations } from '../../i18n'
import type { Department, WorkspaceUser } from '../../types/workspace'

interface DirectoryPanelProps {
  users: WorkspaceUser[]
  departments: Department[]
  query: string
  selectedDepartmentId: number
  currentUserId: number | null
  text: AppTranslations
  loading: boolean
  onQueryChange: (query: string) => void
  onDepartmentChange: (departmentId: number) => void
}

export default function DirectoryPanel({
  users,
  departments,
  query,
  selectedDepartmentId,
  currentUserId,
  text,
  loading,
  onQueryChange,
  onDepartmentChange,
}: DirectoryPanelProps) {
  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.directory.title}</h2>
          <p>{text.directory.subtitle}</p>
        </div>
        <div className="toolbar-search search-shell">
          <select
            value={selectedDepartmentId}
            onChange={(event) => onDepartmentChange(Number(event.target.value))}
            aria-label={text.directory.departmentLabel}
          >
            <option value={0}>{text.directory.allDepartments}</option>
            {departments.map((department) => (
              <option key={department.id} value={department.id}>
                {department.name}
              </option>
            ))}
          </select>
          <input
            type="search"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder={text.directory.searchPlaceholder}
          />
        </div>
      </div>

      <div className="banner-info">{text.directory.searchHint}</div>

      {loading ? (
        <div className="empty-panel">{text.common.loading}</div>
      ) : users.length === 0 ? (
        <div className="empty-panel">{text.directory.noResults}</div>
      ) : (
        <div className="directory-grid">
          {users.map((user) => {
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
                    <dt>{text.directory.employeeNoLabel}</dt>
                    <dd>{user.employee_no || text.common.unavailable}</dd>
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
                    <dd>{user.job_title || text.common.unavailable}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.departmentLabel}</dt>
                    <dd>{user.department_name || text.common.unavailable}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.locationLabel}</dt>
                    <dd>{user.office_location || text.common.unavailable}</dd>
                  </div>
                  <div>
                    <dt>{text.directory.bioLabel}</dt>
                    <dd>{user.bio || text.common.unavailable}</dd>
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
