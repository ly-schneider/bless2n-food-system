import { MenuItem } from '@/types'
import ItemCard from './ItemCard'

interface MenuSectionProps {
  menuItems: MenuItem[]
  stock: Record<string, number>
  onClick: (menuItem: MenuItem) => void
  isLoading: boolean
}

export function MenuSection({ menuItems, stock, onClick, isLoading }: MenuSectionProps) {
  if (isLoading) {
    return (
      <section className="mb-8">
        <h2 className="text-xl font-bold mb-4">Menus</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div
              key={i}
              className="h-32 bg-gray-200 rounded-lg animate-pulse"
            />
          ))}
        </div>
      </section>
    )
  }

  if (!menuItems || menuItems.length === 0) {
    return (
      <section className="mb-8">
        <h2 className="text-xl font-bold mb-4">Menus</h2>
        <p className="text-gray-500">Keine Menus verfügbar</p>
      </section>
    )
  }

  return (
    <section className="mb-8">
      <h2 className="text-xl font-bold mb-4">Menus</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {menuItems.map((menuItem) => (
          <ItemCard
            key={menuItem.id}
            item={menuItem}
            stock={stock[menuItem.id] || 0}
            onClick={() => onClick(menuItem)}
          />
        ))}
      </div>
    </section>
  )
}
