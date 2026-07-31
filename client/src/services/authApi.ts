import { post } from '../utils/requests'

interface LoginCredentials {
  email: string
  password: string
}

interface LoginResponse {
  token: string
  user: {
    id: string
    email: string
    name: string
  }
}

export async function login(credentials: LoginCredentials): Promise<LoginResponse> {
  return post<LoginResponse>('/api/login', credentials)
}
