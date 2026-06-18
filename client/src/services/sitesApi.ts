import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Site } from '../store/sitesSlice'
import { API_BASE_URL } from '../constants/api'

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