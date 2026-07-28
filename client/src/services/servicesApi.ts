import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Service } from '../store/servicesSlice'
import { API_BASE_URL } from '../constants/api'

export const servicesApi = createApi({
  reducerPath: 'servicesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getServices: builder.query<Service[], void>({
      query: () => '/api/services',
    }),
    getServiceById: builder.query<Service, number>({
      query: (id) => `/api/services/${id}`,
    }),
    addService: builder.mutation<Service, Omit<Service, 'id' | 'visits'>>({
      query: (service) => ({
        url: '/api/services',
        method: 'POST',
        body: service,
      }),
    }),
  }),
  tagTypes: ['Service'],
})

export const { useGetServicesQuery, useGetServiceByIdQuery, useAddServiceMutation } = servicesApi
