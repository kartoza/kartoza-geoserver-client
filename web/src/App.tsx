import { useState } from 'react'
import Layout from './components/Layout'
import MainContent from './components/MainContent'
import Dialogs from './components/dialogs'
import { SearchModal, useSearchShortcut } from './components/SearchModal'

function App() {
  const [isSearchOpen, setIsSearchOpen] = useState(false)

  // Enable Ctrl+K global shortcut
  useSearchShortcut(() => setIsSearchOpen(true))

  return (
    <>
      <Layout onSearchClick={() => setIsSearchOpen(true)}>
        <MainContent />
      </Layout>
      <Dialogs />
      <SearchModal
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
      />
    </>
  )
}

export default App
