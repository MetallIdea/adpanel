import { createSlice } from '@reduxjs/toolkit'

export type Service = {
  id: number
  title: string
  description: string
  status: 'active' | 'inactive' | 'pending'
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
