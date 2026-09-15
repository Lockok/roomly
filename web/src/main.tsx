import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './AppVk'
import './styles-vk.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
