import { test, expect, request } from '@playwright/test'

test.describe('Weather API', () => {
  let api: request.APIRequestContext

  test.beforeAll(async ({ playwright }) => {
    api = await playwright.request.newContext({
      baseURL: process.env.APP_BASE_URL || 'http://localhost:8080/',
    })
  })

  test.afterAll(async () => {
    await api.dispose()
  })

  test('GET http://localhost:8080/api/weather should return weather for Kyiv', async () => {
    const res = await api.get('http://localhost:8080/api/weather?city=Kyiv')
    expect(res.status()).toBe(200)
    const json = await res.json()
    expect(json).toHaveProperty('temperature')
    expect(json).toHaveProperty('description')
    expect(json).toHaveProperty('humidity')
  })

  test('POST http://localhost:8080/api/subscribe should accept email and city', async ({ request }) => {
  const email = `test+${Date.now()}@example.com`

  const res = await request.post('http://localhost:8080/api/subscribe', {
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
  test('GET http://localhost:8080/api/confirm/:token should fail with invalid token', async () => {
    const res = await api.get('/api/confirm/invalid-token')
    expect(res.status()).toBeGreaterThanOrEqual(400)
  })

  test('GET http://localhost:8080/api/unsubscribe/:token should fail with invalid token', async () => {
    const res = await api.get('http://localhost:8080/api/unsubscribe/invalid-token')
    expect(res.status()).toBeGreaterThanOrEqual(400)
  })
})
