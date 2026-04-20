import { useState, useEffect } from 'react'

export function useOnlineStatus(url: string): boolean | null {
  const [isOnline, setIsOnline] = useState<boolean | null>(null)

  useEffect(() => {
    if (!url) return

    const check = async () => {
      try {
        const isExternal = url.startsWith('http')
        await fetch(url, {
          method: 'HEAD',
          mode: isExternal ? 'no-cors' : 'cors',
          cache: 'no-store',
          credentials: isExternal ? 'omit' : 'include',
        })
        setIsOnline(true)
      } catch {
        setIsOnline(false)
      }
    }

    check()
    const interval = setInterval(check, 5000)
    return () => clearInterval(interval)
  }, [url])

  return isOnline
}