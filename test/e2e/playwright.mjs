import { chromium } from 'playwright';

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:8080';

let browser;
let page;
let passed = 0;
let failed = 0;

function assert(condition, message) {
  if (condition) {
    console.log(`  ✅ ${message}`);
    passed++;
  } else {
    console.error(`  ❌ ${message}`);
    failed++;
  }
}

async function setup() {
  browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  page = await context.newPage();
}

async function teardown() {
  if (browser) await browser.close();
}

async function testLoginPage() {
  console.log('\n📋 测试：登录页面');
  await page.goto(BASE_URL, { waitUntil: 'networkidle' });

  // 检查页面标题或主要内容
  const title = await page.title();
  assert(title !== '', '页面标题存在');

  // 检查登录表单元素
  const hasUsernameInput = await page.locator('input[type="text"], input[type="email"], input[placeholder*="用户"], input[placeholder*="username" i]').count() > 0;
  const hasPasswordInput = await page.locator('input[type="password"]').count() > 0;
  const hasSubmitButton = await page.locator('button[type="submit"]').count() > 0;

  assert(hasUsernameInput, '用户名输入框存在');
  assert(hasPasswordInput, '密码输入框存在');
  assert(hasSubmitButton, '提交按钮存在');
}

async function testHealthEndpoint() {
  console.log('\n📋 测试：健康检查接口');
  try {
    const resp = await page.request.get(`${API_URL}/health`);
    assert(resp.status() === 200, `健康检查返回 200 (实际: ${resp.status()})`);
    const body = await resp.json();
    assert(body.code === 0 || body.status === 'ok', '健康检查返回正确格式');
  } catch (e) {
    assert(false, `健康检查失败: ${e.message}`);
  }
}

async function testLoginAPI() {
  console.log('\n📋 测试：登录 API');
  try {
    const resp = await page.request.post(`${API_URL}/api/v1/auth/login`, {
      data: { username: 'admin', password: 'admin123' },
      headers: { 'Content-Type': 'application/json' }
    });
    assert(resp.status() === 200 || resp.status() === 401, `登录接口可访问 (状态: ${resp.status()})`);
  } catch (e) {
    assert(false, `登录 API 失败: ${e.message}`);
  }
}

async function testRegisterAPI() {
  console.log('\n📋 测试：注册 API');
  try {
    const resp = await page.request.post(`${API_URL}/api/v1/auth/register`, {
      data: { username: `testuser_${Date.now()}`, password: 'password123' },
      headers: { 'Content-Type': 'application/json' }
    });
    assert(resp.status() === 200 || resp.status() === 409 || resp.status() === 400, `注册接口可访问 (状态: ${resp.status()})`);
  } catch (e) {
    assert(false, `注册 API 失败: ${e.message}`);
  }
}

async function testMetricsEndpoint() {
  console.log('\n📋 测试：Prometheus 监控接口');
  try {
    const resp = await page.request.get(`${API_URL}/metrics`);
    assert(resp.status() === 200, `监控接口返回 200 (实际: ${resp.status()})`);
    const text = await resp.text();
    assert(text.includes('http_requests_total') || text.includes('# HELP'), '返回 Prometheus 指标格式');
  } catch (e) {
    assert(false, `监控接口失败: ${e.message}`);
  }
}

async function run() {
  console.log('🚀 WorkPal E2E 测试开始');
  console.log(`   前端: ${BASE_URL}`);
  console.log(`   后端: ${API_URL}`);

  try {
    await setup();
    await testHealthEndpoint();
    await testMetricsEndpoint();
    await testLoginAPI();
    await testRegisterAPI();
    await testLoginPage();
  } catch (e) {
    console.error('测试过程出错:', e.message);
    failed++;
  } finally {
    await teardown();
  }

  console.log(`\n${'='.repeat(50)}`);
  console.log(`测试结果: ${passed} 通过, ${failed} 失败`);
  process.exit(failed > 0 ? 1 : 0);
}

run();
