import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Site } from '../store/sitesSlice'

const API_BASE_URL = 'http://localhost:3000/api'

export const sitesApi = createApi({
  reducerPath: 'sitesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getSites: builder.query<Site[], void>({
      query: () => '/sites',
    }),
    addSite: builder.mutation<Site, Omit<Site, 'id' | 'visits'>>({
      query: (site) => ({
        url: '/sites',
        method: 'POST',
        body: site,
      }),
    }),
  }),
  tagTypes: ['Site'],
})

export const { useGetSitesQuery, useAddSiteMutation } = sitesApi