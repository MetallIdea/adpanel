import { Menu } from 'primereact/menu'
import type { MenuItem } from 'primereact/menuitem'
import { useNavigate } from 'react-router-dom'
import styles from './Layout.module.css'

export function Layout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  // Меню для левой панели
  const items: MenuItem[] = [
    {
      label: 'Домой',
      icon: 'pi pi-home',
      command: () => navigate('/'),
    },
    {
      label: 'Сайты',
      icon: 'pi pi-globe',
      command: () => navigate('/sites'),
    },
    {
      label: 'Профиль',
      icon: 'pi pi-user',
      command: () => navigate('/profile'),
    },
  ]

  return (
    <div className={styles.root}>
      {/* Левое меню (Sidebar) */}
      <Menu model={items} />

      {/* Основной контент */}
      <main>{children}</main>
    </div>
  )
}