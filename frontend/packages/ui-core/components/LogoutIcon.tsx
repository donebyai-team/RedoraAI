import { FC } from 'react'
import { Tooltip } from '@mui/material'
import { LogOut } from 'lucide-react'
import { useAuth } from '../hooks/useAuth'
import { useIsExecutionRuntimeInExtension } from '../hooks/useExecutionRuntime'

export const LogoutIcon: FC = () => {
  const { logout } = useAuth()

  return (
    <Tooltip title='Log Out' placement='right' arrow disableInteractive>
      <LogOut className='text-back/80 hover:text-back cursor-pointer font-semibold' size={16} onClick={logout} />
    </Tooltip>
  )
}

/**
 * Shows the logout icon only when the execution runtime is in the extension, 'null' otherwise.
 */
export const ExtensionOnlyLogoutIcon: FC = () => {
  const isInExtension = useIsExecutionRuntimeInExtension()

  return isInExtension ? <LogoutIcon /> : null
}
