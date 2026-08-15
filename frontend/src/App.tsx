import { useState } from 'react'
import Sidebar from './features/collections/components/Sidebar'
import { useCollections } from './features/collections/hooks/useCollections'
import LinksPage from './pages/LinksPage'

function App() {
  const [activeCollectionId, setActiveCollectionId] = useState<number | null>(
    null,
  )
  const {
    collections,
    refresh: refreshCollections,
    addCollection,
    editCollection,
    removeCollection,
  } = useCollections()

  async function handleDeleteCollection(id: number) {
    await removeCollection(id)
    if (activeCollectionId === id) {
      setActiveCollectionId(null)
    }
  }

  const activeCollection =
    collections.find((c) => c.id === activeCollectionId) ?? null

  return (
    <div className="flex min-h-screen flex-col bg-stone-50 text-stone-800 sm:flex-row">
      <Sidebar
        collections={collections}
        activeCollectionId={activeCollectionId}
        onSelect={setActiveCollectionId}
        onCreate={async (values, parentId) => {
          await addCollection({ ...values, parent_id: parentId })
        }}
        onEdit={async (id, values) => {
          await editCollection(id, values)
        }}
        onDelete={handleDeleteCollection}
      />
      <div className="flex-1">
        <LinksPage
          activeCollection={activeCollection}
          collections={collections}
          onCollectionsChanged={refreshCollections}
        />
      </div>
    </div>
  )
}

export default App
