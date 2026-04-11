/**
 * Tests for Sidebar component.
 */

import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { render } from '../test/test-utils'
import Sidebar from './Sidebar'

// Mock the ConnectionTree component
vi.mock('./ConnectionTree', () => ({
  default: () => <div data-testid="connection-tree">Connection Tree</div>,
}))

describe('Sidebar', () => {
  it('should render the sidebar', () => {
    render(<Sidebar />)

    expect(screen.getByTestId('connection-tree')).toBeInTheDocument()
  })

  it('should have correct structure', () => {
    const { container } = render(<Sidebar />)

    // Check that the sidebar container exists
    expect(container.firstChild).toBeInTheDocument()
  })
})
