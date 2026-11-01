import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Service } from '../store/servicesSlice'
import { API_BASE_URL } from '../constants/api'

export const servicesApi = createApi({
  reducerPath: 'servicesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getServices: builder.query<Service[], void>({
      query: () => '/api/services',
      providesTags: ['Service'],
    }),
    getServiceById: builder.query<Service, number>({
      query: (id) => `/api/services/${id}`,
      providesTags: ['Service'],
    }),
    addService: builder.mutation<Service, Omit<Service, 'id'>>({
      query: (service) => ({
        url: '/api/services',
        method: 'POST',
        body: service,
      }),
      invalidatesTags: ['Service'],
    }),
    updateService: builder.mutation<Service, Service>({
      query: (service) => ({
        url: `/api/services/${service.id}`,
        method: 'PUT',
        body: service,
      }),
      invalidatesTags: ['Service'],
    }),
    deleteService: builder.mutation<void, number>({
      query: (id) => ({
        url: `/api/services/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Service'],
    }),
    getServicesBySiteId: builder.query<Service[], number>({
      query: (siteId) => `/api/sites/${siteId}/services`,
      providesTags: ['Service'],
    }),
  }),
  tagTypes: ['Service'],
})

export const {
  useGetServicesQuery,
  useGetServiceByIdQuery,
  useAddServiceMutation,
  useUpdateServiceMutation,
  useDeleteServiceMutation,
  useGetServicesBySiteIdQuery,
} = servicesApi
