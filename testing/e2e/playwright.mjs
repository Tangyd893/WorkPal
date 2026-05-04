import { chromium } from '../../frontend/node_modules/playwright/index.mjs'

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
