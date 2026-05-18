import { chromium } from 'playwright'

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000'
const API_URL = process.env.API_URL || 'http://localhost:8080'

const seededAccounts = [
  { username: 'admin', password: 'admin123' },
  { username: 'emma.chen', password: 'workpal123' },
]

let browser
let page
let passed = 0
let failed = 0

function assert(condition, message) {
  if (condition) {
    console.log(`  OK   ${message}`)
    passed++
  } else {
    console.error(`  FAIL ${message}`)
    failed++
  }
}

async function setup() {
  browser = await chromium.launch({ headless: true })
  const context = await browser.newContext()
  page = await context.newPage()
}

async function teardown() {
  if (browser) {
    await browser.close()
  }
}

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

async function loginAPI(account) {
  const response = await page.request.post(`${API_URL}/api/v1/auth/login`, {
    data: account,
    headers: { 'Content-Type': 'application/json' },
  })
  const body = await response.json()
  return { response, body }
}

async function fetchUsers(token) {
  const response = await page.request.get(`${API_URL}/api/v1/users`, {
    headers: authHeaders(token),
  })
  const body = await response.json()
  return body.data?.items ?? body.data ?? []
}

async function testHealthEndpoint() {
  console.log('\n[Test] health endpoint')
  try {
    const response = await page.request.get(`${API_URL}/health`)
    assert(response.status() === 200, `health endpoint returns 200 (actual: ${response.status()})`)
    const body = await response.json()
    assert(body.status === 'ok', 'health payload contains status=ok')
    assert(Boolean(response.headers()['x-request-id']), 'health endpoint returns an X-Request-ID header')
  } catch (error) {
    assert(false, `health endpoint failed: ${error.message}`)
  }
}

async function testGatewayControlPlane() {
  console.log('\n[Test] gateway control plane')

  try {
    const liveResponse = await page.request.get(`${API_URL}/health/live`)
    assert(liveResponse.status() === 200, `gateway live health returns 200 (actual: ${liveResponse.status()})`)
    const liveBody = await liveResponse.json()
    assert(liveBody.service === 'gateway', 'gateway live health identifies the gateway service')

    const readyResponse = await page.request.get(`${API_URL}/health/ready`)
    assert(readyResponse.status() === 200, `gateway readiness returns 200 (actual: ${readyResponse.status()})`)
    const readyBody = await readyResponse.json()
    assert(readyBody.status === 'ok', 'gateway readiness reports status=ok')

    const routesResponse = await page.request.get(`${API_URL}/gateway/routes`)
    assert(routesResponse.status() === 200, `gateway route catalog returns 200 (actual: ${routesResponse.status()})`)
    const routesBody = await routesResponse.json()
    const routes = routesBody.routes ?? []
    assert(Array.isArray(routes) && routes.length > 0, 'gateway route catalog returns route definitions')
    assert(routes.some((route) => route.name === 'gateway-websocket'), 'gateway route catalog includes the websocket route')
    assert(routes.some((route) => route.name === 'workspace-tasks' && route.timeout_ms >= 1), 'gateway route catalog exposes route timeouts')

    const servicesResponse = await page.request.get(`${API_URL}/gateway/services`)
    assert(servicesResponse.status() === 200, `gateway service registry returns 200 (actual: ${servicesResponse.status()})`)
    const servicesBody = await servicesResponse.json()
    const services = servicesBody.services ?? []
    assert(Array.isArray(services) && services.length === 5, 'gateway service registry returns the five downstream services')
    assert(
      services.some((service) => service.name === 'im-service' && service.supports_websocket === true),
      'gateway service registry marks websocket-capable services',
    )
    assert(
      services.some((service) => service.name === 'user-service' && service.circuit_breaker?.state),
      'gateway service registry exposes circuit breaker state',
    )
  } catch (error) {
    assert(false, `gateway control plane failed: ${error.message}`)
  }
}

async function testMetricsEndpoint() {
  console.log('\n[Test] metrics endpoint')
  try {
    const response = await page.request.get(`${API_URL}/metrics`)
    assert(response.status() === 200, `metrics endpoint returns 200 (actual: ${response.status()})`)
    const text = await response.text()
    assert(text.includes('http_requests_total') || text.includes('# HELP'), 'metrics payload looks like Prometheus output')
  } catch (error) {
    assert(false, `metrics endpoint failed: ${error.message}`)
  }
}

async function testSeededLoginsAPI() {
  console.log('\n[Test] seeded login API')

  for (const account of seededAccounts) {
    try {
      const { response, body } = await loginAPI(account)
      assert(response.status() === 200, `${account.username} login returns 200`)
      assert(body.code === 0, `${account.username} login returns code 0`)
      assert(Boolean(body.data?.token), `${account.username} login returns a token`)
      assert(response.headers()['x-upstream-service'] === 'user-service', `${account.username} login is routed through user-service`)
    } catch (error) {
      assert(false, `${account.username} login failed: ${error.message}`)
    }
  }
}

async function testChatAndGroupAPI() {
  console.log('\n[Test] direct and group chat API')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token
    const users = await fetchUsers(token)
    const emma = users.find((user) => user.username === 'emma.chen')
    const liam = users.find((user) => user.username === 'liam.wang')

    assert(Boolean(emma?.id), 'directory API returns Emma Chen')
    assert(Boolean(liam?.id), 'directory API returns Liam Wang')

    const privateResponse = await page.request.post(`${API_URL}/api/v1/conversations`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { type: 1, target_uid: emma.id },
    })
    const privateBody = await privateResponse.json()
    assert(privateResponse.status() === 200 && privateBody.code === 0, 'private conversation can be created')
    assert(privateResponse.headers()['x-upstream-service'] === 'im-service', 'private conversation create is routed through im-service')

    const privateConvID = privateBody.data.id
    const privateMessageResponse = await page.request.post(`${API_URL}/api/v1/conversations/${privateConvID}/messages`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { type: 1, content: 'Private conversation smoke message' },
    })
    const privateMessageBody = await privateMessageResponse.json()
    assert(privateMessageResponse.status() === 200 && privateMessageBody.code === 0, 'private conversation can send messages')

    const groupResponse = await page.request.post(`${API_URL}/api/v1/conversations`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { type: 2, name: `Acceptance Group ${Date.now()}`, member_ids: [emma.id, liam.id] },
    })
    const groupBody = await groupResponse.json()
    assert(groupResponse.status() === 200 && groupBody.code === 0, 'group conversation can be created')

    const groupConvID = groupBody.data.id
    const groupMessageResponse = await page.request.post(`${API_URL}/api/v1/conversations/${groupConvID}/messages`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { type: 1, content: 'Group conversation smoke message' },
    })
    const groupMessageBody = await groupMessageResponse.json()
    assert(groupMessageResponse.status() === 200 && groupMessageBody.code === 0, 'group conversation can send messages')

    const announcementResponse = await page.request.put(`${API_URL}/api/v1/conversations/${groupConvID}/announcement`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { announcement: 'Today we validate announcements and group files.' },
    })
    const announcementBody = await announcementResponse.json()
    assert(announcementResponse.status() === 200 && announcementBody.code === 0, 'group announcement can be updated')

    const uploadResponse = await page.request.post(`${API_URL}/api/v1/files/upload`, {
      headers: authHeaders(token),
      multipart: {
        conv_id: String(groupConvID),
        file: {
          name: 'group-check.txt',
          mimeType: 'text/plain',
          buffer: Buffer.from('group file smoke content', 'utf8'),
        },
      },
    })
    const uploadBody = await uploadResponse.json()
    assert(uploadResponse.status() === 200 && uploadBody.code === 0, 'group file can be uploaded')
    assert(uploadResponse.headers()['x-upstream-service'] === 'file-service', 'group file upload is routed through file-service')

    const filesResponse = await page.request.get(`${API_URL}/api/v1/conversations/${groupConvID}/files`, {
      headers: authHeaders(token),
    })
    const filesBody = await filesResponse.json()
    const files = filesBody.data ?? []
    assert(filesResponse.status() === 200 && Array.isArray(files), 'group files can be listed')
    assert(files.some((file) => file.name === 'group-check.txt'), 'uploaded group file is returned by the API')
  } catch (error) {
    assert(false, `chat and group API smoke failed: ${error.message}`)
  }
}

async function loginUI() {
  await page.goto(BASE_URL, { waitUntil: 'networkidle' })
  await page.locator('#username').fill('admin')
  await page.locator('#password').fill('admin123')
  await page.locator('button[type="submit"]').click()
  await page.waitForURL('**/workspace/overview', { timeout: 15000 })
  await page.getByRole('button', { name: 'English' }).first().click()
}

async function testWorkspaceUI() {
  console.log('\n[Test] workspace UI')

  try {
    await loginUI()
    assert(page.url().includes('/workspace/overview'), 'login redirects to workspace overview')
    assert((await page.getByRole('button', { name: 'Overview' }).count()) > 0, 'language switch updates navigation text')

    await page.getByRole('button', { name: 'Active tasks' }).click()
    await page.waitForURL('**/workspace/tasks', { timeout: 15000 })
    assert(page.url().includes('/workspace/tasks'), 'overview card can jump to the tasks module')

    await page.getByRole('button', { name: 'Directory' }).click()
    await page.waitForURL('**/workspace/directory', { timeout: 15000 })
    await page.locator('.search-shell select').selectOption({ label: 'Engineering' })
    await page.locator('.search-shell input').fill('Platform Engineer')
    await page.waitForTimeout(400)
    assert((await page.locator('text=Liam Wang').count()) > 0, 'directory fuzzy search and department filter find Liam Wang')

    await page.getByRole('button', { name: 'Tasks' }).click()
    await page.getByRole('button', { name: 'Add task' }).click()
    await page.getByLabel('Task title').fill('Acceptance task created in UI')
    await page.getByLabel('Project').fill('Acceptance')
    await page.getByLabel('Summary').fill('Created through the workspace task panel.')
    await page.getByRole('button', { name: 'Create task' }).click()
    assert((await page.locator('text=Acceptance task created in UI').count()) > 0, 'tasks module can create a task')

    await page.getByRole('button', { name: 'Schedule' }).click()
    await page.getByRole('button', { name: 'Add event' }).click()
    await page.getByLabel('Event title').fill('Acceptance schedule event')
    await page.getByLabel('Details').fill('Created through the schedule panel.')
    await page.getByLabel('Starts').fill('2026-05-01T10:00')
    await page.getByLabel('Room').fill('Demo Room')
    await page.getByRole('button', { name: 'Create event' }).click()
    assert((await page.locator('text=Acceptance schedule event').count()) > 0, 'schedule module can create an event')

    await page.getByRole('button', { name: 'Files' }).click()
    await page.locator('input[type="file"]').setInputFiles({
      name: 'acceptance-note.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('acceptance file upload', 'utf8'),
    })
    await page.waitForTimeout(600)
    assert((await page.locator('text=acceptance-note.txt').count()) > 0, 'files module can upload a file')

    await page.getByRole('button', { name: 'Chat' }).click()
    await page.waitForURL('**/workspace/chat', { timeout: 15000 })
    await page.getByRole('button', { name: 'New conversation' }).click()
    await page.locator('.choice-card').filter({ hasText: 'Emma Chen' }).first().click()
    await page.getByRole('button', { name: 'Create' }).click()
    await page.getByPlaceholder('Write a message').fill('UI direct chat smoke message')
    await page.getByRole('button', { name: 'Send' }).click()
    await page.waitForTimeout(500)
    assert((await page.locator('text=UI direct chat smoke message').count()) > 0, 'chat module can create a direct chat and send a message')
  } catch (error) {
    assert(false, `workspace UI failed: ${error.message}`)
  }
}

async function testProjectIssueWorkflow() {
  console.log('\n[Test] Project 2.0 — project creation, issue lifecycle, and kanban')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token

    const projResponse = await page.request.post(`${API_URL}/api/v1/projects`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { key: 'ACCPT', name: 'E2E Acceptance Project', description: 'Project created by E2E smoke test', lead_id: 0, icon: 'folder', category: 'software' },
    })
    const projBody = await projResponse.json()
    assert(projResponse.status() === 200 && projBody.code === 0, 'project can be created')
    const projectID = projBody.data.id
    assert(projectID && projectID.startsWith('prj-'), 'project creation returns a formatted ID')

    const typesResponse = await page.request.get(`${API_URL}/api/v1/projects/${projectID}/issue-types`, {
      headers: authHeaders(token),
    })
    const typesBody = await typesResponse.json()
    const types = typesBody.data ?? []
    assert(Array.isArray(types) && types.length >= 5, 'default issue types are initialized on project creation')

    const taskType = types.find((t) => t.name === 'Task')
    assert(Boolean(taskType), 'default Task issue type exists')

    const issueResponse = await page.request.post(`${API_URL}/api/v1/projects/${projectID}/issues`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: {
        project_id: parseInt(projectID.replace('prj-', ''), 10),
        issue_type_id: taskType?.id ?? 1,
        summary: 'E2E smoke test issue',
        description: 'Created via acceptance test',
        priority: 'High',
        assignee_id: 0,
        reporter_id: 0,
        version_ids: [],
        time_estimate: 0,
      },
    })
    const issueBody = await issueResponse.json()
    assert(issueResponse.status() === 200 && issueBody.code === 0, 'issue can be created under a project')
    const issueID = issueBody.data.id
    assert(issueID && issueID.startsWith('prj-'), 'issue creation returns a formatted ID')
    assert(issueBody.data.status === 'Open', 'new issue starts in Open status')
    assert(issueBody.data.key && issueBody.data.key.startsWith('ACCPT'), 'issue key uses project key prefix')

    const statusResponse = await page.request.put(`${API_URL}/api/v1/issues/${issueID}/status`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { status: 'In Progress' },
    })
    const statusBody = await statusResponse.json()
    assert(statusResponse.status() === 200 && statusBody.code === 0, 'issue status can be transitioned to In Progress')
    assert(statusBody.data.status === 'In Progress', 'issue status updated correctly')

    const changelogResponse = await page.request.get(`${API_URL}/api/v1/issues/${issueID}/changelogs`, {
      headers: authHeaders(token),
    })
    const changelogBody = await changelogResponse.json()
    assert(Array.isArray(changelogBody.data) && changelogBody.data.length > 0, 'issue changelog records are created on status transition')

    const listResponse = await page.request.get(`${API_URL}/api/v1/projects/${projectID}/issues`, {
      headers: authHeaders(token),
    })
    const listBody = await listResponse.json()
    assert(Array.isArray(listBody.data) && listBody.data.length >= 1, 'project issues can be listed')
  } catch (error) {
    assert(false, `project/issue workflow failed: ${error.message}`)
  }
}

async function testDocsAndRevisions() {
  console.log('\n[Test] Docs 2.0 — document CRUD and revisions')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token

    const createResponse = await page.request.post(`${API_URL}/api/v1/documents`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { title: 'E2E Acceptance Document', content: 'Initial revision content', is_folder: false },
    })
    const createBody = await createResponse.json()
    assert(createResponse.status() === 200 && createBody.code === 0, 'document can be created')
    const docID = createBody.data.id
    assert(typeof docID === 'number', 'document creation returns a numeric ID')

    const updateResponse = await page.request.put(`${API_URL}/api/v1/documents/${docID}`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { content: 'Updated revision content' },
    })
    const updateBody = await updateResponse.json()
    assert(updateResponse.status() === 200 && updateBody.code === 0, 'document content can be updated')

    const revisionsResponse = await page.request.get(`${API_URL}/api/v1/documents/${docID}/revisions`, {
      headers: authHeaders(token),
    })
    const revisionsBody = await revisionsResponse.json()
    assert(Array.isArray(revisionsBody.data) && revisionsBody.data.length > 0, 'document revision history is recorded')

    const getResponse = await page.request.get(`${API_URL}/api/v1/documents/${docID}`, {
      headers: authHeaders(token),
    })
    const getBody = await getResponse.json()
    assert(getResponse.status() === 200 && getBody.code === 0, 'document can be retrieved by ID')
    assert(getBody.data.content === 'Updated revision content', 'retrieved document has latest content')

    const listResponse = await page.request.get(`${API_URL}/api/v1/documents`, {
      headers: authHeaders(token),
    })
    const listBody = await listResponse.json()
    assert(Array.isArray(listBody.data) && listBody.data.length >= 1, 'documents list is accessible')
  } catch (error) {
    assert(false, `docs/revisions workflow failed: ${error.message}`)
  }
}

async function testCalendarEventWithAttendees() {
  console.log('\n[Test] Calendar 2.0 — event CRUD with attendees')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token
    const users = await fetchUsers(token)
    const emma = users.find((u) => u.username === 'emma.chen')

    const createResponse = await page.request.post(`${API_URL}/api/v1/calendar`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: {
        title: 'E2E Calendar Event',
        description: 'Acceptance test calendar event',
        starts_at: '2026-05-20T09:00:00Z',
        ends_at: '2026-05-20T10:00:00Z',
        is_all_day: false,
        location: 'Meeting Room A',
        attendee_ids: emma ? [emma.id] : [],
      },
    })
    const createBody = await createResponse.json()
    assert(createResponse.status() === 200 && createBody.code === 0, 'calendar event can be created')
    const eventID = createBody.data.id
    assert(typeof eventID === 'number', 'calendar event creation returns a numeric ID')

    const getResponse = await page.request.get(`${API_URL}/api/v1/calendar/${eventID}`, {
      headers: authHeaders(token),
    })
    const getBody = await getResponse.json()
    assert(getResponse.status() === 200 && getBody.code === 0, 'calendar event can be retrieved')
    assert(getBody.data.title === 'E2E Calendar Event', 'retrieved event has correct title')

    if (emma) {
      assert(Array.isArray(getBody.data.attendees) && getBody.data.attendees.length > 0, 'calendar event has attendees')
    }

    const updateResponse = await page.request.put(`${API_URL}/api/v1/calendar/${eventID}`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { description: 'Updated description' },
    })
    const updateBody = await updateResponse.json()
    assert(updateResponse.status() === 200 && updateBody.code === 0, 'calendar event can be updated')

    const deleteResponse = await page.request.delete(`${API_URL}/api/v1/calendar/${eventID}`, {
      headers: authHeaders(token),
    })
    assert(deleteResponse.status() === 200, 'calendar event can be deleted')
  } catch (error) {
    assert(false, `calendar workflow failed: ${error.message}`)
  }
}

async function testApprovalSubmitAndAction() {
  console.log('\n[Test] Approval 2.0 — template, submission, approval/rejection')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token

    const templateResponse = await page.request.post(`${API_URL}/api/v1/approvals/templates`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { name: 'E2E Approval Template', description: 'Acceptance test approval template', form_schema: '{"fields":[]}' },
    })
    const templateBody = await templateResponse.json()
    assert(templateResponse.status() === 200 && templateBody.code === 0, 'approval template can be created')
    const templateID = templateBody.data.id

    const instanceResponse = await page.request.post(`${API_URL}/api/v1/approvals/instances`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { template_id: templateID, title: 'E2E Approval Request', form_data: '{}' },
    })
    const instanceBody = await instanceResponse.json()
    assert(instanceResponse.status() === 200 && instanceBody.code === 0, 'approval instance can be submitted')
    const instanceID = instanceBody.data.id

    const getResponse = await page.request.get(`${API_URL}/api/v1/approvals/instances/${instanceID}`, {
      headers: authHeaders(token),
    })
    const getBody = await getResponse.json()
    assert(getResponse.status() === 200 && getBody.code === 0, 'approval instance can be retrieved')

    const actionResponse = await page.request.post(`${API_URL}/api/v1/approvals/instances/${instanceID}/action`, {
      headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
      data: { action: 'approve', comment: 'Approved by E2E test' },
    })
    const actionBody = await actionResponse.json()
    assert(actionResponse.status() === 200 && actionBody.code === 0, 'approval instance can be processed with approve action')

    const listResponse = await page.request.get(`${API_URL}/api/v1/approvals/instances`, {
      headers: authHeaders(token),
    })
    const listBody = await listResponse.json()
    assert(Array.isArray(listBody.data) && listBody.data.length >= 1, 'approval instances list is accessible')
  } catch (error) {
    assert(false, `approval workflow failed: ${error.message}`)
  }
}

async function testNotificationIntegration() {
  console.log('\n[Test] Notification 2.0 — list, unread count, mark read')

  try {
    const { body } = await loginAPI({ username: 'admin', password: 'admin123' })
    const token = body.data.token

    const listResponse = await page.request.get(`${API_URL}/api/v1/notifications`, {
      headers: authHeaders(token),
    })
    const listBody = await listResponse.json()
    assert(listResponse.status() === 200 && listBody.code === 0, 'notifications list endpoint is accessible')
    assert(Array.isArray(listBody.data), 'notifications list returns an array')

    const unreadResponse = await page.request.get(`${API_URL}/api/v1/notifications/unread-count`, {
      headers: authHeaders(token),
    })
    const unreadBody = await unreadResponse.json()
    assert(unreadResponse.status() === 200 && unreadBody.code === 0, 'unread notification count is accessible')
    assert(typeof unreadBody.data?.count === 'number', 'unread count returns a numeric value')

    const markAllResponse = await page.request.put(`${API_URL}/api/v1/notifications/read-all`, {
      headers: authHeaders(token),
    })
    assert(markAllResponse.status() === 200, 'mark all notifications read succeeds')
  } catch (error) {
    assert(false, `notification integration failed: ${error.message}`)
  }
}

async function run() {
  console.log('WorkPal E2E smoke test')
  console.log(`  frontend: ${BASE_URL}`)
  console.log(`  backend : ${API_URL}`)

  try {
    await setup()
    await testHealthEndpoint()
    await testGatewayControlPlane()
    await testMetricsEndpoint()
    await testSeededLoginsAPI()
    await testChatAndGroupAPI()
    await testWorkspaceUI()
    await testProjectIssueWorkflow()
    await testDocsAndRevisions()
    await testCalendarEventWithAttendees()
    await testApprovalSubmitAndAction()
    await testNotificationIntegration()
  } catch (error) {
    console.error(`Unexpected test error: ${error.message}`)
    failed++
  } finally {
    await teardown()
  }

  console.log(`\nResult: ${passed} passed, ${failed} failed`)
  process.exit(failed > 0 ? 1 : 0)
}

run()
