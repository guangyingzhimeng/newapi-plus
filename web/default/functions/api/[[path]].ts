const DEFAULT_BACKEND_BASE = 'https://guangyingzhimeng.dpdns.org/new-api'

type Env = {
  BACKEND_BASE_URL?: string
}

function backendBase(env: Env): string {
  return (env.BACKEND_BASE_URL || DEFAULT_BACKEND_BASE).replace(/\/+$/, '')
}

function copyRequestHeaders(request: Request): Headers {
  const headers = new Headers(request.headers)
  headers.delete('host')
  headers.delete('cf-connecting-ip')
  headers.delete('cf-ipcountry')
  headers.delete('cf-ray')
  headers.delete('cf-visitor')
  headers.set('x-forwarded-host', new URL(request.url).host)
  headers.set('x-forwarded-proto', 'https')
  return headers
}

function proxiedURL(request: Request, env: Env): string {
  const url = new URL(request.url)
  const path = url.pathname.replace(/^\/api\/?/, '/api/')
  return `${backendBase(env)}${path}${url.search}`
}

export async function onRequest(context: {
  request: Request
  env: Env
}): Promise<Response> {
  const { request, env } = context
  const headers = copyRequestHeaders(request)
  const method = request.method.toUpperCase()

  const response = await fetch(proxiedURL(request, env), {
    method,
    headers,
    body: method === 'GET' || method === 'HEAD' ? undefined : request.body,
    redirect: 'manual',
  })

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: response.headers,
  })
}
