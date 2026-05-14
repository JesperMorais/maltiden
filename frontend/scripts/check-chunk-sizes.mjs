#!/usr/bin/env node
import { readFileSync, readdirSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { join, resolve } from 'node:path'

const MAX_GZ_BYTES = 90 * 1024 // 90 KB

const distDir = resolve(import.meta.dirname, '../dist/assets')

let files
try {
  files = readdirSync(distDir).filter((f) => f.endsWith('.js'))
} catch {
  console.error(`dist/assets not found — run 'npm run build' first`)
  process.exit(1)
}

const rows = files.map((file) => {
  const raw = readFileSync(join(distDir, file))
  const gz = gzipSync(raw)
  return { file, rawKB: (raw.length / 1024).toFixed(1), gzKB: (gz.length / 1024).toFixed(1), gzBytes: gz.length }
})

rows.sort((a, b) => b.gzBytes - a.gzBytes)

const pad = (s, n) => String(s).padStart(n)
console.log(`${'File'.padEnd(60)} ${'Raw KB'.padStart(8)} ${'GZ KB'.padStart(8)} ${'Limit'.padStart(8)}`)
console.log('-'.repeat(86))

let failures = 0
for (const { file, rawKB, gzKB, gzBytes } of rows) {
  const over = gzBytes > MAX_GZ_BYTES
  if (over) failures++
  const flag = over ? ' ❌ OVER LIMIT' : ''
  console.log(`${file.padEnd(60)} ${pad(rawKB, 8)} ${pad(gzKB, 8)} ${pad('90', 8)}${flag}`)
}

console.log()
if (failures > 0) {
  console.error(`${failures} chunk(s) exceed 90 KB gzipped`)
  process.exit(1)
} else {
  console.log(`All ${rows.length} chunks ≤ 90 KB gzipped ✓`)
}
