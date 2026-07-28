import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { File } from '../store/filesSlice'
import { API_BASE_URL } from '../constants/api'

export const filesApi = createApi({
  reducerPath: 'filesApi',
  baseQuery: fetchBaseQuery({ baseUrl: API_BASE_URL }),
  endpoints: (builder) => ({
    getFiles: builder.query<File[], void>({
      query: () => 'api/files',
    }),
  }),
  tagTypes: ['File'],
})

export const { useGetFilesQuery } = filesApi
