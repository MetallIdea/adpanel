import { post } from '../utils/requests'

interface LoginCredentials {
  login: string
  password: string
}

interface LoginResponse {
  token: string
}

export async function login(credentials: LoginCredentials): Promise<LoginResponse> {
  return post<LoginResponse>('/login', credentials)
}
