import http from 'k6/http'
import { check, sleep } from 'k6'
import { Trend, Rate } from 'k6/metrics'

export const options = {
  scenarios: {
    两千人群发消息: {
      executor: 'ramping-vus',
      stages: [
        { duration: '30s', target: 200 },
        { duration: '2m', target: 2000 },
        { duration: '1m', target: 2000 },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    '消息发送耗时': ['p(99)<1500'],
  },
}

const 发送耗时 = new Trend('消息发送耗时')
const 发送失败率 = new Rate('消息发送失败率')

const 网关地址 = __ENV.WORKPAL_GATEWAY || 'http://localhost:8080'
const 会话编号 = Number(__ENV.WORKPAL_CONV_ID || '1')
const 用户名 = __ENV.WORKPAL_USERNAME || 'admin'
const 密码 = __ENV.WORKPAL_PASSWORD || 'admin123'

function 登录() {
  const 响应 = http.post(
    `${网关地址}/api/v1/auth/login`,
    JSON.stringify({ username: 用户名, password: 密码 }),
    { headers: { 'Content-Type': 'application/json' } },
  )
  check(响应, { 登录成功: (res) => res.status === 200 })
  const 结果 = 响应.json()
  return 结果?.data?.token || 结果?.token || ''
}

export function setup() {
  return { token: 登录() }
}

export default function (data) {
  const 幂等键 = `k6-${__VU}-${__ITER}-${Date.now()}`
  const 响应 = http.post(
    `${网关地址}/api/v1/conversations/${会话编号}/messages`,
    JSON.stringify({
      type: 1,
      content: `群发压测消息 vu=${__VU} iter=${__ITER}`,
      idempotency_key: 幂等键,
    }),
    {
      headers: {
        Authorization: `Bearer ${data.token}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': 幂等键,
        'X-Trace-ID': `perf-${__VU}-${__ITER}`,
      },
    },
  )

  发送耗时.add(响应.timings.duration)
  发送失败率.add(响应.status >= 400)
  check(响应, { 消息发送成功: (res) => res.status === 200 })
  sleep(0.2)
}
