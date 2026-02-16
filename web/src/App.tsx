import { useState, useCallback } from 'react'
import Layout from './components/Layout'
import MainContent from './components/MainContent'
import Dialogs from './components/dialogs'
import { SearchModal, useSearchShortcut } from './components/SearchModal'
import { HelpPanel, useHelpShortcut } from './components/HelpPanel'

function App() {
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [isHelpOpen, setIsHelpOpen] = useState(false)

  // Enable Ctrl+K global shortcut
  useSearchShortcut(() => setIsSearchOpen(true))

  // Enable ? global shortcut for help
  const toggleHelp = useCallback(() => setIsHelpOpen(prev => !prev), [])
  useHelpShortcut(toggleHelp)

  return (
    <>
      <Layout
        onSearchClick={() => setIsSearchOpen(true)}
        onHelpClick={() => setIsHelpOpen(true)}
      >
        <MainContent />
      </Layout>
      <Dialogs />
      <SearchModal
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
      />
      <HelpPanel
        isOpen={isHelpOpen}
        onClose={() => setIsHelpOpen(false)}
      />
    </>
  )
}

export default App
