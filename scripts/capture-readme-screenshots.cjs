const fs = require('node:fs/promises')
const path = require('node:path')
const { chromium } = require('playwright')
const sharp = require('sharp')

const root = path.resolve(__dirname, '..')
const outDir = path.join(root, 'docs', 'screenshots')
const baseUrl = 'http://127.0.0.1:5173/dashboard'
const now = new Date('2026-08-22T07:30:00Z')

const isoAgo = (minutes) => new Date(now.getTime() - minutes * 60 * 1000).toISOString()

const user = {
  id: 'op-01',
  tenant_id: 'tenant-demo',
  name: 'Operator Demo',
  email: 'operator@example.test',
  role: 'ADMIN',
  is_active: true,
  totp_verified_at: isoAgo(4000),
}

const operators = [
  user,
  {
    id: 'op-02',
    tenant_id: 'tenant-demo',
    name: 'Support Lead',
    email: 'lead@example.test',
    role: 'OPERATOR',
    is_active: true,
  },
]

const accounts = [
  {
    id: 'acct-01',
    host_id: 'demo-channel-01',
    display_name: 'Main Support',
    health: 'healthy',
    is_connected: true,
    queue_size: 2,
  },
  {
    id: 'acct-02',
    host_id: 'demo-channel-02',
    display_name: 'Sales Desk',
    health: 'healthy',
    is_connected: true,
    queue_size: 0,
  },
  {
    id: 'acct-03',
    host_id: 'demo-channel-03',
    display_name: 'After Hours',
    health: 'disconnected',
    is_connected: false,
    queue_size: 6,
  },
]

const conversationBase = {
  tenant_id: 'tenant-demo',
  account_id: 'acct-01',
  contact_id: 'contact-demo',
  contact_number: '+10000000000',
  ticket_number: 1248,
  status: 'HANDED_OFF',
  bot_state: 'handoff_requested',
  started_at: isoAgo(180),
  last_activity_at: isoAgo(4),
  closed_at: null,
  handoff_at: isoAgo(22),
  closure_reason: null,
  assignee: 'Support Lead',
  merged_into_id: null,
  created_at: isoAgo(180),
  updated_at: isoAgo(4),
}

const conversations = [
  {
    ...conversationBase,
    id: 'conv-01',
    contact_name: 'Customer A',
    contact_number: '+10000000001',
    is_group: false,
    last_message_preview: 'Thanks, the team can follow up after confirmation.',
    last_message_actor: 'CONTACT',
  },
  {
    ...conversationBase,
    id: 'conv-02',
    ticket_number: 1247,
    contact_name: 'Customer Group',
    contact_number: '+10000000002',
    is_group: true,
    status: 'OPEN',
    assignee: 'Operator Demo',
    last_activity_at: isoAgo(17),
    last_message_preview: 'Can you share availability for tomorrow?',
    last_message_actor: 'CONTACT',
  },
  {
    ...conversationBase,
    id: 'conv-03',
    ticket_number: 1246,
    contact_name: 'Customer B',
    contact_number: '+10000000003',
    status: 'WAITING',
    assignee: null,
    last_activity_at: isoAgo(42),
    last_message_preview: 'We sent the latest estimate and are waiting.',
    last_message_actor: 'OPERATOR',
  },
]

const selectedConversation = {
  ...conversations[0],
  contact_email: 'hidden@example.test',
  contact_location: 'Metro Area',
}

const messages = [
  {
    id: 'msg-01',
    tenant_id: 'tenant-demo',
    conversation_id: 'conv-01',
    actor: 'CONTACT',
    provider: 'whatsmeow',
    provider_message_id: 'provider-01',
    direction: 'INCOMING',
    sender_address: null,
    content: 'Hi, can your team help with our order update?',
    message_type: 'TEXT',
    media_url: '',
    status: 'RECEIVED',
    provider_timestamp: isoAgo(62),
    is_internal: false,
    created_at: isoAgo(62),
    updated_at: isoAgo(62),
  },
  {
    id: 'msg-02',
    tenant_id: 'tenant-demo',
    conversation_id: 'conv-01',
    actor: 'BOT',
    provider: 'whatsmeow',
    provider_message_id: 'provider-02',
    direction: 'OUTGOING',
    content: 'Thanks for reaching out. I’ll collect the details and bring in a teammate.',
    message_type: 'TEXT',
    media_url: '',
    status: 'SENT',
    provider_timestamp: isoAgo(60),
    is_internal: false,
    created_at: isoAgo(60),
    updated_at: isoAgo(60),
  },
  {
    id: 'msg-03',
    tenant_id: 'tenant-demo',
    conversation_id: 'conv-01',
    actor: 'SYSTEM',
    provider: 'whatsmeow',
    provider_message_id: 'provider-03',
    direction: 'OUTGOING',
    content: 'Internal note: customer prefers a concise update before noon.',
    message_type: 'TEXT',
    media_url: '',
    status: 'NOTE',
    provider_timestamp: isoAgo(35),
    is_internal: true,
    created_at: isoAgo(35),
    updated_at: isoAgo(35),
  },
  {
    id: 'msg-04',
    tenant_id: 'tenant-demo',
    conversation_id: 'conv-01',
    actor: 'OPERATOR',
    operator_id: 'op-02',
    operator_name: 'Support Lead',
    provider: 'whatsmeow',
    provider_message_id: 'provider-04',
    direction: 'OUTGOING',
    content: 'We checked the request and will confirm the final slot shortly.',
    message_type: 'TEXT',
    media_url: '',
    status: 'DELIVERED',
    provider_timestamp: isoAgo(21),
    is_internal: false,
    created_at: isoAgo(21),
    updated_at: isoAgo(21),
  },
  {
    id: 'msg-05',
    tenant_id: 'tenant-demo',
    conversation_id: 'conv-01',
    actor: 'CONTACT',
    provider: 'whatsmeow',
    provider_message_id: 'provider-05',
    direction: 'INCOMING',
    content: 'Thanks, the team can follow up after confirmation.',
    message_type: 'TEXT',
    media_url: '',
    status: 'RECEIVED',
    provider_timestamp: isoAgo(4),
    is_internal: false,
    created_at: isoAgo(4),
    updated_at: isoAgo(4),
  },
]

function metrics() {
  return {
    inbound: 186,
    outbound: 171,
    failed: 3,
    status_breakdown: { SENT: 84, DELIVERED: 61, READ: 26 },
    buckets: Array.from({ length: 24 }, (_, i) => ({
      start: isoAgo((24 - i) * 60),
      inbound: 3 + ((i * 7) % 18),
      outbound: 2 + ((i * 5) % 16),
      failed: i % 9 === 0 ? 1 : 0,
    })),
  }
}

function queueDepth() {
  return Array.from({ length: 36 }, (_, i) => ({
    id: `q-${i}`,
    host_id: 'demo-channel-01',
    queue_size: Math.max(0, Math.round(2 + Math.sin(i / 3) * 2 + (i % 10 === 0 ? 4 : 0))),
    timestamp: isoAgo(36 - i),
  }))
}

async function fulfillJson(route, data) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(data),
    headers: { 'Access-Control-Allow-Origin': '*' },
  })
}

async function anonymize(input, output) {
  const img = sharp(input)
  const meta = await img.metadata()
  await img
    .composite([
      {
        input: await sharp(input).blur(7).toBuffer(),
        blend: 'over',
        left: 0,
        top: 0,
      },
    ])
    .resize({ width: Math.min(meta.width || 1440, 1440), withoutEnlargement: true })
    .jpeg({ quality: 88, mozjpeg: true })
    .toFile(output)
}

async function main() {
  await fs.mkdir(outDir, { recursive: true })

  const browser = await chromium.launch({
    headless: true,
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  })
  const page = await browser.newPage({ viewport: { width: 1440, height: 960 }, deviceScaleFactor: 1 })

  await page.route('**/*', async (route) => {
    const req = route.request()
    const url = new URL(req.url())

    if (req.method() === 'OPTIONS') {
      return route.fulfill({ status: 204, headers: { 'Access-Control-Allow-Origin': '*' } })
    }
    if (url.pathname === '/dashboard/api/me') {
      return fulfillJson(route, { user, tenant_id: 'tenant-demo', tenant_slug: 'demo-workspace', tenant_name: 'Demo Workspace' })
    }
    if (url.pathname === '/dashboard/api/accounts') {
      return fulfillJson(route, accounts)
    }
    if (url.pathname === '/dashboard/api/operators') {
      return fulfillJson(route, operators)
    }
    if (url.pathname === '/api/v1/conversations') {
      return fulfillJson(route, conversations)
    }
    if (url.pathname === '/api/v1/conversations/conv-01') {
      return fulfillJson(route, { conversation: selectedConversation, messages })
    }
    if (url.pathname === '/dashboard/api/monitoring/status') {
      return fulfillJson(route, {
        host_id: 'demo-channel-01',
        status: 'ONLINE',
        is_connected: true,
        uptime: '18h 42m',
        queue_size: 2,
        last_connected_at: isoAgo(1122),
        last_disconnected_at: isoAgo(1660),
        updated_at: isoAgo(1),
      })
    }
    if (url.pathname === '/dashboard/api/monitoring/metrics') {
      return fulfillJson(route, metrics())
    }
    if (url.pathname === '/dashboard/api/monitoring/queue-depth') {
      return fulfillJson(route, queueDepth())
    }
    if (url.pathname === '/dashboard/api/monitoring/stream') {
      return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    }

    return route.continue()
  })

  await page.addInitScript(() => {
    window.localStorage.setItem('whatsapp_dashboard_tenant_id', 'tenant-demo')
    window.localStorage.setItem('whatsapp_dashboard_tenant_slug', 'demo-workspace')
    window.localStorage.setItem('conversation-details-collapsed', 'false')
    window.localStorage.setItem('sidebar-collapsed', 'false')
  })

  const rawInbox = path.join(outDir, 'whops-dashboard-inbox.raw.png')
  const rawMonitoring = path.join(outDir, 'whops-dashboard-monitoring.raw.png')
  const inbox = path.join(outDir, 'whops-dashboard-inbox.jpg')
  const monitoring = path.join(outDir, 'whops-dashboard-monitoring.jpg')

  await page.goto(`${baseUrl}/conversations/conv-01`, { waitUntil: 'networkidle' })
  await page.waitForSelector('text=Internal note', { timeout: 15_000 })
  await page.screenshot({ path: rawInbox, fullPage: true })
  await anonymize(rawInbox, inbox)

  await page.goto(`${baseUrl}/monitoring`, { waitUntil: 'networkidle' })
  await page.waitForSelector('text=Message Volume', { timeout: 15_000 })
  await page.screenshot({ path: rawMonitoring, fullPage: true })
  await anonymize(rawMonitoring, monitoring)

  await fs.rm(rawInbox, { force: true })
  await fs.rm(rawMonitoring, { force: true })
  await browser.close()
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
