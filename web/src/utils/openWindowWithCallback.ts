/**
 * Opens a URL in a new window and calls `onClose` when the window is closed
 * or when the child window sends a `checkout_status` postMessage signal.
 * Uses the parent window's `focus` event to detect closure (no polling).
 */
export function openWindowWithCallback(url: string, onClose: () => void): void {
  const newWindow = window.open(url, '_blank')
  if (!newWindow) return

  const cleanup = () => {
    window.removeEventListener('focus', onFocus)
    window.removeEventListener('message', onMessage)
  }

  const onMessage = (event: MessageEvent) => {
    if (event.source === newWindow && event.data?.type === 'checkout_status') {
      cleanup()
      onClose()
    }
  }

  const onFocus = () => {
    if (newWindow.closed) {
      cleanup()
      onClose()
    }
  }

  window.addEventListener('message', onMessage)
  window.addEventListener('focus', onFocus)
}