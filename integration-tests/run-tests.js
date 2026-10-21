// Starts orgviz with -configdir and runs mocha.

import { createServer } from 'node:net'
import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import waitOn from 'wait-on'

const __dirname = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(__dirname, '..')
const serviceBin = join(repoRoot, 'service', 'orgviz')
const frontendDist = join(repoRoot, 'frontend', 'dist')
const configDir = join(repoRoot, 'var', 'config-skel')

async function getFreePort () {
  return new Promise((resolve, reject) => {
    const server = createServer()
    server.listen(0, '127.0.0.1', () => {
      const address = server.address()
      const port = typeof address === 'object' && address ? address.port : 0
      server.close((err) => (err ? reject(err) : resolve(port)))
    })
    server.on('error', reject)
  })
}

async function main () {
  if (!existsSync(serviceBin)) {
    throw new Error(`orgviz binary not found at ${serviceBin}. Run "make service-build" first.`)
  }

  if (!existsSync(join(frontendDist, 'index.html'))) {
    throw new Error(`Frontend build not found at ${frontendDist}. Run "make frontend-build" first.`)
  }

  const testPort = process.env.ORGVIZ_TEST_PORT
    ? Number(process.env.ORGVIZ_TEST_PORT)
    : await getFreePort()
  const baseUrl = process.env.ORGVIZ_BASE_URL || `http://127.0.0.1:${testPort}`

  const env = {
    ...process.env,
    PORT: String(testPort),
    ORGVIZ_FRONTEND_DIR: frontendDist
  }

  const child = spawn(serviceBin, ['-configdir', configDir], {
    cwd: join(repoRoot, 'service'),
    env,
    stdio: ['ignore', 'inherit', 'inherit']
  })

  try {
    await waitOn({
      resources: [`http-get://127.0.0.1:${testPort}/`],
      timeout: 30000,
      validateStatus: (status) => status >= 200 && status < 500
    })

    const mocha = spawn(
      'npx',
      ['mocha', '--timeout', '30000', 'tests/**/*.spec.js'],
      {
        cwd: __dirname,
        env: {
          ...env,
          ORGVIZ_BASE_URL: baseUrl
        },
        stdio: 'inherit'
      }
    )

    const code = await new Promise((resolve) => mocha.on('close', resolve))
    child.kill('SIGTERM')
    process.exit(code ?? 1)
  } catch (err) {
    child.kill('SIGTERM')
    throw err
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
