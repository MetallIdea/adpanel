import { createSlice } from '@reduxjs/toolkit'

export type File = {
  id: number
  name: string
  type: 'file' | 'folder'
  size: number
  path: string
  createdAt: string
}

export type FilesState = {
}

const initialState: FilesState = {
}

const filesSlice = createSlice({
  name: 'files',
  initialState,
  reducers: {
  },
})

export const {
} = filesSlice.actions

export const filesReducer = filesSlice.reducer
