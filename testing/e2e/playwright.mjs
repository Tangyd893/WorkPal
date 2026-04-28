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
    console.log(`  OK  ${message}`)
    passed++
  } else {
    console.error(`  FAIL  ${message}`)
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

async function testHealthEndpoint() {
  console.log('\n[Test] health endpoint')
  try {
    const response = await page.request.get(`${API_URL}/health`)
    assert(response.status() === 200, `health endpoint returns 200 (actual: ${response.status()})`)
    const body = await response.json()
    assert(body.status === 'ok', 'health payload contains status=ok')
  } catch (error) {
    assert(false, `health endpoint failed: ${error.message}`)
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
      const response = await page.request.post(`${API_URL}/api/v1/auth/login`, {
        data: account,
        headers: {
          'Content-Type': 'application/json',
        },
      })

      assert(response.status() === 200, `${account.username} login returns 200`)
      const body = await response.json()
      assert(body.code === 0, `${account.username} login returns code 0`)
      assert(Boolean(body.data?.token), `${account.username} login returns a token`)
    } catch (error) {
      assert(false, `${account.username} login failed: ${error.message}`)
    }
  }
}

async function testWorkspaceUI() {
  console.log('\n[Test] workspace UI')

  await page.goto(BASE_URL, { waitUntil: 'networkidle' })
  const title = await page.title()
  assert(title !== '', 'page title exists')

  const usernameInput = page.locator('#username')
  const passwordInput = page.locator('#password')
  await usernameInput.fill('admin')
  await passwordInput.fill('admin123')
  await page.locator('button[type="submit"]').click()

  await page.waitForURL('**/workspace/overview', { timeout: 15000 })
  assert(page.url().includes('/workspace/overview'), 'login redirects to workspace overview')

  const navLabels = ['总览', '沟通', '任务', '日程', '文件', '通讯录']
  for (const label of navLabels) {
    const visible = (await page.getByRole('button', { name: label }).count()) > 0
    assert(visible, `navigation shows ${label}`)
  }

  await page.getByRole('button', { name: 'English' }).click()
  assert((await page.getByRole('button', { name: 'Overview' }).count()) > 0, 'language switch updates navigation text')

  await page.getByRole('button', { name: 'Directory' }).click()
  await page.waitForURL('**/workspace/directory', { timeout: 15000 })
  assert((await page.locator('text=@admin').count()) > 0, 'directory renders the admin account')
  assert((await page.locator('text=emma.chen@workpal.local').count()) > 0, 'directory renders seeded employee data')

  await page.getByRole('button', { name: 'Chat' }).click()
  await page.waitForURL('**/workspace/chat', { timeout: 15000 })
  assert((await page.getByRole('button', { name: 'New conversation' }).count()) > 0, 'chat module renders create conversation action')
  await page.getByRole('button', { name: 'New conversation' }).click()
  assert((await page.locator('text=emma.chen').count()) > 0, 'conversation modal lists seeded teammates')
}

async function run() {
  console.log('WorkPal E2E smoke test')
  console.log(`  frontend: ${BASE_URL}`)
  console.log(`  backend : ${API_URL}`)

  try {
    await setup()
    await testHealthEndpoint()
    await testMetricsEndpoint()
    await testSeededLoginsAPI()
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
