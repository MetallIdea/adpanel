import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Site } from '../store/sitesSlice'
import { API_BASE_URL } from '../constants/api'

export const sitesApi = createApi({
  reducerPath: 'sitesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getSites: builder.query<Site[], void>({
      query: () => '/sites',
      providesTags: ['Site'],
    }),
    getSiteById: builder.query<Site, number>({
      query: (id) => `/sites/${id}`,
      providesTags: ['Site'],
    }),
    addSite: builder.mutation<Site, Omit<Site, 'id' | 'visits'>>({
      query: (site) => ({
        url: '/sites',
        method: 'POST',
        body: site,
      }),
      invalidatesTags: ['Site'],
    }),
    updateSite: builder.mutation<Site, Site>({
      query: (site) => ({
        url: `/sites/${site.id}`,
        method: 'PUT',
        body: site,
      }),
      invalidatesTags: ['Site'],
    }),
    getNginxConfig: builder.query<{ nginx_config: string }, number>({
      query: (id) => `/sites/${id}/nginx-config`,
    }),
    updateNginxConfig: builder.mutation<{ message: string }, { id: number; nginx_config: string }>({
      query: ({ id, nginx_config }) => ({
        url: `/sites/${id}/nginx-config`,
        method: 'PUT',
        body: { nginx_config },
      }),
    }),
  }),
  tagTypes: ['Site'],
})

export const {
  useGetSitesQuery,
  useGetSiteByIdQuery,
  useAddSiteMutation,
  useUpdateSiteMutation,
  useGetNginxConfigQuery,
  useUpdateNginxConfigMutation,
} = sitesApi
