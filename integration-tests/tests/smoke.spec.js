import assert from 'node:assert'
import { Builder, By, until } from 'selenium-webdriver'
import chrome from 'selenium-webdriver/chrome.js'

const base = process.env.ORGVIZ_BASE_URL || 'http://127.0.0.1:18080'

describe('orgviz smoke', function () {
  this.timeout(30000)
  let driver

  before(async function () {
    const options = new chrome.Options()
    options.addArguments('--headless=new', '--no-sandbox', '--disable-dev-shm-usage')
    driver = await new Builder().forBrowser('chrome').setChromeOptions(options).build()
  })

  after(async function () {
    if (driver) {
      await driver.quit()
    }
  })

  it('renders the Example Company chart', async function () {
    await driver.get(base + '/')
    await driver.wait(until.elementLocated(By.css('#picocrank-sidebar')), 10000)
    const nav = await driver.findElement(By.css('#picocrank-sidebar')).getText()
    assert.match(nav, /Example Company/)
    await driver.wait(until.elementLocated(By.css('div.person')), 10000)
    const people = await driver.findElements(By.css('div.person'))
    assert.ok(people.length >= 3)
    const body = await driver.findElement(By.css('body')).getText()
    assert.match(body, /Alice Smith/)
  })

  it('switches to Second Company from the sidebar', async function () {
    await driver.get(base + '/')
    await driver.wait(until.elementLocated(By.css('#picocrank-sidebar a')), 10000)
    const links = await driver.findElements(By.css('#picocrank-sidebar a'))
    const second = await Promise.all(links.map(async (link) => {
      const text = await link.getText()
      return text.includes('Second Company') ? link : null
    }))
    const target = second.find(Boolean)
    assert.ok(target, 'Second Company link')
    await target.click()
    await driver.wait(async () => {
      const text = await driver.findElement(By.css('body')).getText()
      return text.includes('Founder') || text.includes('CFO')
    }, 10000)
  })

  it('serves prometheus metrics', async function () {
    const response = await fetch(base + '/metrics')
    assert.equal(response.status, 200)
    const body = await response.text()
    assert.match(body, /go_goroutines|promhttp/)
  })
})
