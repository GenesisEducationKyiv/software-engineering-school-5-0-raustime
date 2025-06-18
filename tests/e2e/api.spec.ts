import { test, expect, request } from '@playwright/test'

test.describe('Weather API', () => {
  let api: request.APIRequestContext

  test.beforeAll(async ({ playwright }) => {
    api = await playwright.request.newContext({
      baseURL: process.env.APP_BASE_URL || 'http://api:8080',
    })
  })

  test.afterAll(async () => {
    await api.dispose()
  })

  test('GET /api/weather should return weather for Kyiv', async () => {
    const res = await api.get('/api/weather?city=Kyiv')
    expect(res.status()).toBe(200)
    const json = await res.json()
    expect(json).toHaveProperty('temperature')
    expect(json).toHaveProperty('description')
    expect(json).toHaveProperty('humidity')
  })

  test('POST /api/subscribe should accept email and city', async ({ request }) => {
  const email = `test+${Date.now()}@example.com`

  const res = await request.post('/api/subscribe', {
  headers: {
    'Content-Type': 'application/json',
  },
  data: {
    email: `test+${Date.now()}@example.com`,
    city: 'Kyiv',
    frequency: 'daily',  // <-- обов’язкове поле!
  },
})
expect(res.status()).toBe(200)
const text = await res.text()
})

  // Ці два тести — умовні, бо токен буде реальний лише після email (мокати або інтегрувати з БД)
  test('GET /api/confirm/:token should fail with invalid token', async () => {
    const res = await api.get('/api/confirm/invalid-token')
    expect(res.status()).toBeGreaterThanOrEqual(400)
  })

  test('GET /api/unsubscribe/:token should fail with invalid token', async () => {
    const res = await api.get('/api/unsubscribe/invalid-token')
    expect(res.status()).toBeGreaterThanOrEqual(400)
  })
})
