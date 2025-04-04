import React, {  useState } from 'react'
import {  Menu, MenuItem, Tooltip } from '@mui/material'
import {  ChevronDown, LogOut } from 'lucide-react'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { isPlatformAdmin } from '@doota/ui-core/helper/role'
import { useOrganization } from '@doota/ui-core/hooks/useOrganization'
import toast from 'react-hot-toast'
import { errorToMessage } from '@doota/pb/utils/errors'

const NavBar: React.FC<{ hoverActive: boolean }> = ({ hoverActive }) => {
  const { user, logout } = useAuth()
  const [currentOrg, setCurrentOrganization] = useOrganization()
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const open = Boolean(anchorEl)

  const getInitials = (name: string | undefined) => {
    const matches = name?.match(/[A-Z]/g)
    const numUpperCase = matches ? matches.length : 0
    const className = numUpperCase >= 3 ? 'text-base ml-[-4px]' : 'text-xl'
    return { initials: matches ? matches.join('') : '', className }
  }

  const handleOpenOrg = (event: React.MouseEvent<HTMLDivElement>) => {
    if (user && isPlatformAdmin(user) && user.organizations.length > 1) {
      setAnchorEl(event.currentTarget)
    }
  }
  const handleClose = () => {
    setAnchorEl(null)
  }

  const handleLogout = () => {
    logout()
  }

  const canChangeOrg = user && isPlatformAdmin(user) && user.organizations.length > 1

  return (
    <>
      <div className='text-white flex font-extrabold px-2 h-14 border-b border-neutral-800 items-center'>
        <div className={`${hoverActive ? 'hidden' : ''} ${getInitials(currentOrg?.name).className}`}>
          {getInitials(currentOrg?.name).initials}
        </div>
        <div
          role='presentation'
          className={`${
            hoverActive ? 'flex' : 'hidden'
          } justify-between w-full relative text-xl truncate h-full items-center`}
          onClick={canChangeOrg ? handleOpenOrg : undefined}
        >
          <Tooltip
            title='Select your organization'
            placement='right'
            arrow
            disableInteractive
            disableHoverListener={canChangeOrg ? false : true}
          >
            <div className='flex items-center gap-[4px] text-white/60 hover:text-white'>
              <div className='text-white'>{currentOrg?.name}</div>
              {canChangeOrg && (
                <div className={`${hoverActive ? 'inline-block' : 'hidden'}`}>
                  <ChevronDown size={18} />
                </div>
              )}
            </div>
          </Tooltip>
          <Tooltip title='Log Out' placement='right' arrow disableInteractive>
            <LogOut className='text-white/80 hover:text-white cursor-pointer' size={14} onClick={handleLogout} />
          </Tooltip>
        </div>
        {user && isPlatformAdmin(user) && (
          <Menu
            id='basic-menu'
            anchorEl={anchorEl}
            open={open}
            onClose={handleClose}
            MenuListProps={{
              'aria-labelledby': 'basic-button'
            }}
          >
            {user.organizations.map((organization, index) => (
              <MenuItem
                key={index}
                onClick={() => {
                  setCurrentOrganization(organization)
                    .then(() => {
                      window.location.reload()
                    })
                    .catch(error => toast.error(errorToMessage(error)))
                }}
              >
                {organization.name}
              </MenuItem>
            ))}
          </Menu>
        )}
      </div>

      {/* Navigation */}
    </>
  )
}

export default NavBar
