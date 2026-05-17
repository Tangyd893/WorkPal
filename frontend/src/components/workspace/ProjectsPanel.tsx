import { useState } from 'react'
import type { AppTranslations } from '../../i18n'
import type { CreateIssueInput, CreateProjectInput, Issue, IssueType, Project } from '../../types/workspace'

interface Props {
  projects: Project[]
  issues: Issue[]
  issueTypes: IssueType[]
  selectedProjectId: string | null
  issuesLoading: boolean
  text: AppTranslations
  getDisplayName: (username: string) => string
  onSelectProject: (projectId: string) => void
  onAddProject: (draft: CreateProjectInput) => void
  onDeleteProject: (projectId: string) => void
  onAddIssue: (projectId: string, draft: CreateIssueInput) => void
  onUpdateIssueStatus: (issueId: string, status: string) => void
  onDeleteIssue: (issueId: string) => void
}

const PRIORITIES = ['Critical', 'High', 'Medium', 'Low'] as const
const KANBAN_COLUMNS = [
  { status: 'Open', labelKey: 'backlog' as const },
  { status: 'In Progress', labelKey: 'inProgress' as const },
  { status: 'In Review', labelKey: 'inReview' as const },
  { status: 'Done', labelKey: 'done' as const },
]

export default function ProjectsPanel({ projects, issues, issueTypes, selectedProjectId, issuesLoading, text, getDisplayName, onSelectProject, onAddProject, onDeleteProject, onAddIssue, onUpdateIssueStatus, onDeleteIssue }: Props) {
  const [showAddProject, setShowAddProject] = useState(false)
  const [showAddIssue, setShowAddIssue] = useState(false)
  const [projectForm, setProjectForm] = useState<CreateProjectInput>({ key: '', name: '', description: '', lead_id: 0, icon: 'folder', category: 'software' })
  const [issueForm, setIssueForm] = useState<CreateIssueInput>({ project_id: 0, issue_type_id: 0, parent_id: null, summary: '', description: '', priority: 'Medium', assignee_id: null, reporter_id: 0, due_date: null, story_points: null, version_ids: [], time_estimate: 0 })
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null)

  const selectedProject = projects.find((p) => p.id === selectedProjectId)
  const defaultType = issueTypes.length > 0 ? issueTypes[0].id : 0

  const handleDragStart = (e: React.DragEvent, issueId: string) => {
    e.dataTransfer.setData('text/plain', issueId)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleDragOver = (e: React.DragEvent, status: string) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverColumn(status)
  }

  const handleDragLeave = () => setDragOverColumn(null)

  const handleDrop = (e: React.DragEvent, status: string) => {
    e.preventDefault()
    setDragOverColumn(null)
    const issueId = e.dataTransfer.getData('text/plain')
    if (issueId) onUpdateIssueStatus(issueId, status)
  }

  return (
    <section className="projects-panel">
      <header className="panel-header">
        <div>
          <h2>{text.projects.title}</h2>
          <p>{text.projects.subtitle}</p>
        </div>
        <button className="primary-button" onClick={() => { setShowAddProject(!showAddProject); setShowAddIssue(false) }}>
          + {text.projects.addProject}
        </button>
      </header>

      {showAddProject && (
        <form className="add-project-form" onSubmit={(e) => { e.preventDefault(); onAddProject(projectForm); setShowAddProject(false); setProjectForm({ key: '', name: '', description: '', lead_id: 0, icon: 'folder', category: 'software' }) }}>
          <div className="form-row">
            <input className="text-input compact" placeholder={text.projects.projectKey} value={projectForm.key} onChange={(e) => setProjectForm({ ...projectForm, key: e.target.value })} required />
            <input className="text-input compact" placeholder={text.projects.projectName} value={projectForm.name} onChange={(e) => setProjectForm({ ...projectForm, name: e.target.value })} required />
            <button className="primary-button" type="submit">{text.projects.createProject}</button>
            <button className="secondary-button" type="button" onClick={() => setShowAddProject(false)}>{text.common.cancel}</button>
          </div>
        </form>
      )}

      <div className="projects-layout">
        {projects.length === 0 ? (
          <div className="empty-panel">{text.projects.emptyProjects}</div>
        ) : (
          <div className="project-list-sidebar">
            {projects.map((project) => (
              <button
                key={project.id}
                className={selectedProjectId === project.id ? 'project-card active' : 'project-card'}
                onClick={() => { onSelectProject(project.id); setShowAddIssue(false) }}
              >
                <div className="project-card-header">
                  <span className="project-key">{project.key}</span>
                  <span className="project-name">{project.name}</span>
                </div>
                <button className="icon-button danger" onClick={(e) => { e.stopPropagation(); onDeleteProject(project.id) }} title={text.projects.deleteProject}>×</button>
              </button>
            ))}
          </div>
        )}

        {selectedProject && (
          <div className="project-board">
            <header className="board-header">
              <h3>{selectedProject.name} ({selectedProject.key})</h3>
              <button className="primary-button compact" onClick={() => { const tid = defaultType; setIssueForm({ ...issueForm, project_id: parseInt(selectedProject.id.replace('prj-', '')) || 0, issue_type_id: tid }); setShowAddIssue(!showAddIssue) }}>
                + {text.projects.addIssue}
              </button>
            </header>

            {showAddIssue && (
              <form className="add-issue-form" onSubmit={(e) => {
                e.preventDefault()
                if (!selectedProject) return
                onAddIssue(selectedProject.id, issueForm)
                setShowAddIssue(false)
                setIssueForm({ project_id: 0, issue_type_id: 0, parent_id: null, summary: '', description: '', priority: 'Medium', assignee_id: null, reporter_id: 0, due_date: null, story_points: null, version_ids: [], time_estimate: 0 })
              }}>
                <div className="form-row">
                  <input className="text-input compact" placeholder={text.projects.issueSummary} value={issueForm.summary} onChange={(e) => setIssueForm({ ...issueForm, summary: e.target.value })} required />
                  <select className="text-input compact" value={issueForm.issue_type_id || defaultType} onChange={(e) => setIssueForm({ ...issueForm, issue_type_id: parseInt(e.target.value) })}>
                    {issueTypes.map((t) => (<option key={t.id} value={t.id}>{t.name}</option>))}
                  </select>
                  <select className="text-input compact" value={issueForm.priority} onChange={(e) => setIssueForm({ ...issueForm, priority: e.target.value })}>
                    {PRIORITIES.map((p) => (<option key={p} value={p}>{p}</option>))}
                  </select>
                  <button className="primary-button compact" type="submit">{text.projects.createIssue}</button>
                  <button className="secondary-button compact" type="button" onClick={() => setShowAddIssue(false)}>{text.common.cancel}</button>
                </div>
              </form>
            )}

            {issuesLoading ? (
              <div className="skeleton-panel">{text.common.loading}</div>
            ) : (
              <div className="kanban-board">
                {KANBAN_COLUMNS.map((col) => {
                  const colIssues = issues.filter((iss) => iss.status === col.status)
                  return (
                    <div
                      key={col.status}
                      className={`kanban-column ${dragOverColumn === col.status ? 'drag-over' : ''}`}
                      onDragOver={(e) => handleDragOver(e, col.status)}
                      onDragLeave={handleDragLeave}
                      onDrop={(e) => handleDrop(e, col.status)}
                    >
                      <div className="kanban-column-header">
                        <span className="column-title">{text.projects[col.labelKey]}</span>
                        <span className="column-count">{colIssues.length}</span>
                      </div>
                      <div className="kanban-cards">
                        {colIssues.map((issue) => (
                          <div
                            key={issue.id}
                            className="issue-card"
                            draggable
                            onDragStart={(e) => handleDragStart(e, issue.id)}
                          >
                            <div className="issue-card-key">{issue.key}</div>
                            <div className="issue-card-summary">{issue.summary}</div>
                            <div className="issue-card-meta">
                              {issue.issue_type_name && <span className="issue-badge">{issue.issue_type_name}</span>}
                              <span className={`priority-badge ${issue.priority.toLowerCase()}`}>{issue.priority}</span>
                              {issue.assignee_id && <span className="assignee-badge">{getDisplayName(String(issue.assignee_id))}</span>}
                            </div>
                            <button className="icon-button danger compact" onClick={() => onDeleteIssue(issue.id)} title={text.projects.deleteIssue}>×</button>
                          </div>
                        ))}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  )
}
