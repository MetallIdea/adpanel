import { createSlice } from '@reduxjs/toolkit'

export type Service = {
  id: number
  site_id: number
  name: string
  port: number
  working_dir: string
  startup_cmd: string
  env_content: string
  comment: string
}

export type ServicesState = {
}

const initialState: ServicesState = {
}

const servicesSlice = createSlice({
  name: 'services',
  initialState,
  reducers: {
  },
})

export const {
} = servicesSlice.actions

export const servicesReducer = servicesSlice.reducer
