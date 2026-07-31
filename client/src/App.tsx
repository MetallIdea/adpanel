import { Routes, Route, Navigate } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { Layout } from './components/Layout/Layout'

const Home = lazy(() => import('./pages/Home/Home'))
const Sites = lazy(() => import('./pages/Sites/Sites'))
const CreateSite = lazy(() => import('./pages/Sites/CreateSite'))
const Services = lazy(() => import('./pages/Services/Services'))
const CreateService = lazy(() => import('./pages/Services/CreateService'))
const ServiceDetails = lazy(() => import('./pages/Services/ServiceDetails'))
const Files = lazy(() => import('./pages/Files/Files'))
const Login = lazy(() => import('./pages/Login/Login'))

export default function App() {
  return (
    <Layout>
      <Suspense fallback={<div>Loading...</div>}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<Home />} />
          <Route path="/sites" element={<Sites />} />
          <Route path="/sites/create" element={<CreateSite />} />
          <Route path="/services" element={<Services />} />
          <Route path="/services/create" element={<CreateService />} />
          <Route path="/services/:id" element={<ServiceDetails />} />
          <Route path="/files" element={<Files />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </Layout>
  )
}
