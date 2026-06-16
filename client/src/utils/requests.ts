const API_BASE_URL = 'http://localhost:3000/api'

interface RequestOptions extends RequestInit {
  headers?: Record<string, string>
}

async function checkResponse(response: Response) {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
  }
  return response
}

export async function get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })
  await checkResponse(response)
  return response.json()
}

export async function post<T>(endpoint: string, body?: unknown, options?: RequestOptions): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    body: body ? JSON.stringify(body) : undefined,
    ...options,
  })
  await checkResponse(response)
  return response.json()
}

export async function put<T>(endpoint: string, body?: unknown, options?: RequestOptions): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const response = await fetch(url, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    body: body ? JSON.stringify(body) : undefined,
    ...options,
  })
  await checkResponse(response)
  return response.json()
}

export async function del<T>(endpoint: string, options?: RequestOptions): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const response = await fetch(url, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })
  await checkResponse(response)
  return response.json()
}
