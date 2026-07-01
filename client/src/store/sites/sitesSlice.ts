import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

export interface Site {
  id: number
  name: string
  url: string
  status: 'active' | 'inactive' | 'pending'
  visits: number
}

export interface SitesState {
  sites: Site[]
  loading: boolean
  error: string | null
}

const initialState: SitesState = {
  sites: [],
  loading: false,
  error: null,
}

const sitesSlice = createSlice({
  name: 'sites',
  initialState,
  reducers: {
    fetchSitesRequest(state) {
      state.loading = true
      state.error = null
    },
    fetchSitesSuccess(state, action: PayloadAction<Site[]>) {
      state.sites = action.payload
      state.loading = false
    },
    fetchSitesFailure(state, action: PayloadAction<string>) {
      state.error = action.payload
      state.loading = false
    },
    addSite(state, action: PayloadAction<Site>) {
      state.sites.push(action.payload)
    },
    updateSite(state, action: PayloadAction<Site>) {
      const index = state.sites.findIndex(site => site.id === action.payload.id)
      if (index !== -1) {
        state.sites[index] = action.payload
      }
    },
    deleteSite(state, action: PayloadAction<number>) {
      state.sites = state.sites.filter(site => site.id !== action.payload)
    },
  },
})

export const {
  fetchSitesRequest,
  fetchSitesSuccess,
  fetchSitesFailure,
  addSite,
  updateSite,
  deleteSite,
} = sitesSlice.actions

export const sitesReducer = sitesSlice.reducer
