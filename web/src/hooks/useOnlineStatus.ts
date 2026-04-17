import { useState, useEffect } from 'react'

export function useOnlineStatus(url: string): boolean | null {
  const [isOnline, setIsOnline] = useState<boolean | null>(null)

  useEffect(() => {
    if (!url) return

    const check = async () => {
      try {
        await fetch(url, { method: 'HEAD', mode: 'no-cors', cache: 'no-store' })
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