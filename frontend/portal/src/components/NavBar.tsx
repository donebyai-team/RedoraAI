import React, { useState } from 'react'
import { Button, Menu, MenuItem, Tooltip } from '@mui/material'
import { ChevronDown, LogOut, Settings } from 'lucide-react'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { isPlatformAdmin, isAdmin } from '@doota/ui-core/helper/role'
import { useOrganization } from '@doota/ui-core/hooks/useOrganization'
import { routes } from '@doota/ui-core/routing'
import toast from 'react-hot-toast'
import { errorToMessage } from '@doota/pb/utils/errors'

import {
  Box,
  Typography,
  Slider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  ListItemButton,
  Avatar,
  Divider,
  IconButton,
  Badge,
  Paper,
} from "@mui/material"
import {
  Mail as MailIcon,
  MoreVert as MoreVertIcon,
  Add as AddIcon,
  Chat as ChatIcon,
  Settings as SettingsIcon,
} from "@mui/icons-material"

const NavBar: React.FC<{ hoverActive?: boolean }> = ({ hoverActive }) => {
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

  const canChangeOrg = user && isPlatformAdmin(user) && user.organizations.length > 1;

  const [relevancy, setRelevancy] = useState<number>(40)

  const handleRelevancyChange = (_event: Event, newValue: number | number[]) => {
    setRelevancy(newValue as number)
  }

  return (
    <>
      {/* <div className='text-white flex font-extrabold px-2 h-14 border-b border-neutral-800 items-center'>
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
                      <ChevronDown size={18}/>
                    </div>
                )}
              </div>
            </Tooltip>
            <Tooltip title='Log Out' placement='right' arrow disableInteractive>
              <LogOut className='text-white/80 hover:text-white cursor-pointer' size={14} onClick={handleLogout}/>
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
        </div> */}

      {/* Navigation */}
      {/* <div className='flex flex-col justify-top h-[calc(100%_-_56px)] gap-1 p-1'>
          {user && isAdmin(user) && (
              <Button
                  startIcon={<Settings className='w-2 text-neutral-500'/>}
                  variant='contained'
                  className={`!bg-transparent !shadow-none relative hover:!bg-neutral-900 !min-w-full w-full h-[40px] !justify-start items-center !normal-case ${
                      hoverActive ? 'flex' : 'hidden'
                  }`}
                  href={routes.app.settings.account}
              >
                <div className={`${hoverActive ? 'flex' : 'hidden'} text-neutral-200 hover:text-neutral-50 font-bold`}>
                  Settings
                </div>
              </Button>
          )}
        </div> */}

      <Paper
        elevation={0}
        sx={{
          width: 280,
          height: "100vh",
          borderRight: "1px solid #e0e0e0",
          borderRadius: 0,
          bgcolor: "white",
        }}
      >
        <Box sx={{ p: 2, display: "flex", flexDirection: "column", height: "100%" }}>
          {/* Avatar */}
          <Box sx={{ display: "flex", justifyContent: "flex-start", mb: 2 }}>
            <Avatar sx={{ width: 48, height: 48, bgcolor: "#f3f4f6", color: "#111827" }}>A</Avatar>
          </Box>

          {/* Inbox */}
          <ListItem
            component={Box}
            sx={{
              bgcolor: "#F5F5F5",
              borderRadius: 1,
              mb: 3,
              p: "8px 16px",
              "&:hover": { bgcolor: "#EFEFEF" },
            }}
          >
            <ListItemIcon>
              <MailIcon color="action" />
            </ListItemIcon>
            <ListItemText
              primary={
                <Typography variant="body1" fontWeight={500}>
                  {`Inbox`}
                </Typography>
              }
            />
            <Badge
              badgeContent={2}
              color="warning"
              sx={{
                "& .MuiBadge-badge": {
                  bgcolor: "#FF9800",
                  color: "white",
                  fontWeight: "bold",
                },
              }}
            />
          </ListItem>

          {/* Filters */}
          <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 2 }}>
            FILTERS
          </Typography>

          <Box sx={{ mb: 3 }}>
            <Box sx={{ display: "flex", justifyContent: "space-between", mb: 2 }}>
              <Typography variant="body2">Relevancy</Typography>
              <Typography variant="body2" color="text.secondary">
                {relevancy}%
              </Typography>
            </Box>
            <Slider
              value={relevancy}
              onChange={handleRelevancyChange}
              sx={{
                color: "#FF9800",
                "& .MuiSlider-thumb": {
                  width: 16,
                  height: 16,
                },
              }}
            />
          </Box>

          {/* Filter by Subreddit */}
          <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 2 }}>
            FILTER BY SUBREDDIT
          </Typography>

          <List sx={{ p: 0, mb: "auto" }}>
            <ListItem
              disablePadding
              secondaryAction={
                <Badge
                  badgeContent={43}
                  color="warning"
                  sx={{
                    "& .MuiBadge-badge": {
                      bgcolor: "#FF9800",
                      color: "white",
                      fontWeight: "bold",
                    },
                  }}
                />
              }
              sx={{ mb: 2 }}
            >
              <ListItemButton sx={{ borderRadius: 1 }}>
                <ListItemText primary="r/marketing" />
              </ListItemButton>
            </ListItem>

            <ListItem
              disablePadding
              secondaryAction={
                <Box sx={{ display: "flex", alignItems: "center" }}>
                  <Badge
                    badgeContent={141}
                    color="warning"
                    sx={{
                      mr: 1,
                      "& .MuiBadge-badge": {
                        bgcolor: "#FF9800",
                        color: "white",
                        fontWeight: "bold",
                      },
                    }}
                  />
                  <IconButton edge="end" size="small">
                    <MoreVertIcon fontSize="small" />
                  </IconButton>
                </Box>
              }
            >
              <ListItemButton sx={{ borderRadius: 1, bgcolor: "#F0F5FF" }}>
                <ListItemText primary="r/sales" />
              </ListItemButton>
            </ListItem>
          </List>

          {/* Bottom Actions */}
          <Box sx={{ mt: 2 }}>
            <Divider sx={{ mb: 2 }} />
            <List sx={{ p: 0 }}>
              <ListItem disablePadding>
                <ListItemButton sx={{ borderRadius: 1 }}>
                  <ListItemIcon sx={{ minWidth: 40 }}>
                    <AddIcon />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Typography variant="body2" fontWeight={500}>
                        Add subreddits
                      </Typography>
                    }
                  />
                </ListItemButton>
              </ListItem>

              <ListItem disablePadding>
                <ListItemButton sx={{ borderRadius: 1 }}>
                  <ListItemIcon sx={{ minWidth: 40 }}>
                    <SettingsIcon />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Typography variant="body2" fontWeight={500}>
                        Settings
                      </Typography>
                    }
                  />
                </ListItemButton>
              </ListItem>
            </List>
          </Box>
        </Box>
      </Paper>
    </>
  )
}

export default NavBar
