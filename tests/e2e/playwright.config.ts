import { defineConfig } from '@playwright/test';

export default defineConfig({
  use: {
    baseURL: process.env.APP_BASE_URL || 'http://api:8080',
  },
});
