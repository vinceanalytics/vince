const { test } = require('./support/harness')
const { mockRequest, mockManyRequests, isMac, expectCustomEvent } = require('./support/test-utils')
const { expect } = require('@playwright/test')

test.describe('tagged-events extension', () => {
    test('tracks a tagged link click with custom props + url prop', async ({ page }) => {
        await page.goto('/tagged-event.html')

        const linkURL = await page.locator('#link').getAttribute('href')

        const vinceRequestMock = mockRequest(page, '/api/event')
        await page.click('#link')
        expectCustomEvent(await vinceRequestMock, 'Payment Complete', { amount: '100', method: "Credit Card", url: linkURL })
    });

    test('tracks a tagged form submit with custom props when submitting by pressing enter', async ({ page }) => {
        await page.goto('/tagged-event.html')
        const vinceRequestMock = mockRequest(page, '/api/event')

        const inputLocator = page.locator('#form-text-input')
        await inputLocator.type('some input')
        await inputLocator.press('Enter')

        expectCustomEvent(await vinceRequestMock, 'Signup', { type: "Newsletter" })
    });

    test('tracks submit on a form with a tagged parent when submit button is clicked', async ({ page }) => {
        await page.goto('/tagged-event.html')

        const vinceRequestMockList = mockManyRequests(page, '/api/event', 2)

        await page.click('#submit-form-with-tagged-parent')

        const requests = await vinceRequestMockList

        expect(requests.length).toBe(1)
        expectCustomEvent(requests[0], "Form Submit", {})
    });

    test('tracks click and auxclick on any tagged HTML element', async ({ page }, workerInfo) => {
        await page.goto('/tagged-event.html')

        const vinceRequestMockList = mockManyRequests(page, '/api/event', 3)

        await page.click('#button')
        await page.click('#span')
        await page.click('#div', { modifiers: [isMac(workerInfo) ? 'Meta' : 'Control'] })

        const requests = await vinceRequestMockList
        expect(requests.length).toBe(3)
        requests.forEach(request => expectCustomEvent(request, 'Custom Event', { foo: "bar" }))
    });

    test('does not track elements without vince-event-name class + link elements navigate', async ({ page }) => {
        await page.goto('/tagged-event.html')
        const linkLocator = page.locator('#not-tracked-link')

        const linkURL = await linkLocator.getAttribute('href')

        const vinceRequestMock = mockRequest(page, '/api/event')
        const navigationRequestMock = mockRequest(page, linkURL)

        await page.click('#not-tracked-button')
        await page.click('#not-tracked-span')
        await linkLocator.click()

        expect(await vinceRequestMock, "should not have made vince request").toBeNull()
        expect((await navigationRequestMock).url()).toContain(linkURL)
    });

    test('tracks tagged HTML elements when their child element is clicked', async ({ page }, workerInfo) => {
        await page.goto('/tagged-event.html')

        const vinceRequestMockList = mockManyRequests(page, '/api/event', 2)

        await page.click('#h2-with-link-parent', { modifiers: [isMac(workerInfo) ? 'Meta' : 'Control'] })
        await page.click('#link-with-div-parent')

        const requests = await vinceRequestMockList
        expect(requests.length).toBe(2)
        requests.forEach(request => expectCustomEvent(request, 'Custom Event', { foo: "bar" }))
    });

    test('tracks tagged element that is dynamically added to the DOM', async ({ page }) => {
        await page.goto('/tagged-event.html')

        const vinceRequestMock = mockRequest(page, '/api/event')

        const buttonLocator = page.locator('#dynamic-tagged-button')
        await buttonLocator.waitFor({ state: 'visible' })
        await page.waitForTimeout(500)

        await buttonLocator.click()

        expectCustomEvent(await vinceRequestMock, 'Custom Event', {})
    });

    test('does not track clicks inside a tagged form, except submit click', async ({ page }) => {
        await page.goto('/tagged-event.html')

        const vinceRequestMock = mockRequest(page, '/api/event')

        await page.click('#form')
        await page.click('#form-input')
        await page.click('#form-div')

        expect(await vinceRequestMock, "should not have made vince request").toBeNull()
    });
});
