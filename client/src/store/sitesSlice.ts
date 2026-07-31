import { createSlice } from '@reduxjs/toolkit'

export type Site = {
  id: number
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
  visits: number
}

export type SitesState = {
}

const initialState: SitesState = {
}

const sitesSlice = createSlice({
  name: 'sites',
  initialState,
  reducers: {
  },
})

export const {
} = sitesSlice.actions

export const sitesReducer = sitesSlice.reducer
