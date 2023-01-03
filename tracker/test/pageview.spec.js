const { test } = require('./support/harness');
const { mockRequest } = require('./support/test-utils')
const { expect } = require('@playwright/test');

test.describe('Basic installation', () => {
  test('Sends pageview automatically', async ({ page }) => {
    const vinceRequestMock = mockRequest(page, '/api/event')
    await page.goto('/simple.html');

    const vinceRequest = await vinceRequestMock;
    expect(vinceRequest.url()).toContain('/api/event')
    expect(vinceRequest.postDataJSON().n).toEqual('pageview')
  });
});
