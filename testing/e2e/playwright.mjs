import { chromium } from '../../frontend/node_modules/playwright/index.mjs'

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000'
const API_URL = process.env.API_URL || 'http://localhost:8080'

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

async function createTestUser() {
  const suffix = Date.now()
  const username = `playwright_${suffix}`
  const password = 'pass123456'
  const email = `${username}@example.com`

  const response = await page.request.post(`${API_URL}/api/v1/auth/register`, {
    data: {
      username,
      password,
      nickname: 'Playwright User',
      email,
    },
    headers: {
      'Content-Type': 'application/json',
    },
  })

  assert(response.status() === 200, `register API returns 200 (actual: ${response.status()})`)

  const body = await response.json()
  assert(body.code === 0, 'register API returns code 0')

  return { username, password }
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

async function testLoginAPI(credentials) {
  console.log('\n[Test] login API')
  try {
    const response = await page.request.post(`${API_URL}/api/v1/auth/login`, {
      data: credentials,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    assert(response.status() === 200, `login API returns 200 (actual: ${response.status()})`)
    const body = await response.json()
    assert(body.code === 0, 'login API returns code 0')
    assert(Boolean(body.data?.token), 'login API returns a token')
  } catch (error) {
    assert(false, `login API failed: ${error.message}`)
  }
}

async function testLoginPage() {
  console.log('\n[Test] login page')
  await page.goto(BASE_URL, { waitUntil: 'networkidle' })

  const title = await page.title()
  assert(title !== '', 'page title exists')

  const hasUsernameInput = (await page.locator('input[type="text"], input[type="email"], #username').count()) > 0
  const hasPasswordInput = (await page.locator('input[type="password"], #password').count()) > 0
  const hasSubmitButton = (await page.locator('button[type="submit"]').count()) > 0

  assert(hasUsernameInput, 'username input exists')
  assert(hasPasswordInput, 'password input exists')
  assert(hasSubmitButton, 'submit button exists')
}

async function run() {
  console.log('WorkPal E2E smoke test')
  console.log(`  frontend: ${BASE_URL}`)
  console.log(`  backend : ${API_URL}`)

  try {
    await setup()
    await testHealthEndpoint()
    await testMetricsEndpoint()
    const credentials = await createTestUser()
    await testLoginAPI(credentials)
    await testLoginPage()
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
