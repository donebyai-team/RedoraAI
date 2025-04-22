import React, { useEffect, useState } from 'react'
// import { Button, Menu, MenuItem, Tooltip } from '@mui/material'
// import { ChevronDown, LogOut, Settings } from 'lucide-react'
// import { useAuth } from '@doota/ui-core/hooks/useAuth'
// import { isPlatformAdmin, isAdmin } from '@doota/ui-core/helper/role'
// import { useOrganization } from '@doota/ui-core/hooks/useOrganization'
// import { routes } from '@doota/ui-core/routing'
// import toast from 'react-hot-toast'
// import { errorToMessage } from '@doota/pb/utils/errors'
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
  Menu,
  MenuItem,
} from "@mui/material"
import {
  Mail as MailIcon,
  MoreVert as MoreVertIcon,
  Add as AddIcon,
  Settings as SettingsIcon,
  DeleteOutline,
} from "@mui/icons-material"
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import AddSubredditDialog from './AddSubredditDialog';

const SUBREDDIT_LIST = [
  {
    prefix: "r/",
    name: "marketing",
    badge_count: 64
  },
  {
    prefix: "r/",
    name: "sales",
    badge_count: 90
  }
];

const SIDEBAR_MENU_LIST = [
  {
    name: "inbox",
    active_menu: "inbox"
  },
  {
    name: "marketing",
    active_menu: "marketing"
  },
  {
    name: "sales",
    active_menu: "sales"
  },
];

const NavBar: React.FC<{ hoverActive?: boolean }> = ({ hoverActive }) => {
  // const { user, logout } = useAuth()
  // const [currentOrg, setCurrentOrganization] = useOrganization()
  // const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  // const open = Boolean(anchorEl)

  // const getInitials = (name: string | undefined) => {
  //   const matches = name?.match(/[A-Z]/g)
  //   const numUpperCase = matches ? matches.length : 0
  //   const className = numUpperCase >= 3 ? 'text-base ml-[-4px]' : 'text-xl'
  //   return { initials: matches ? matches.join('') : '', className }
  // }

  // const handleOpenOrg = (event: React.MouseEvent<HTMLDivElement>) => {
  //   if (user && isPlatformAdmin(user) && user.organizations.length > 1) {
  //     setAnchorEl(event.currentTarget)
  //   }
  // }
  // const handleClose = () => {
  //   setAnchorEl(null)
  // }

  // const handleLogout = () => {
  //   logout()
  // }

  // const canChangeOrg = user && isPlatformAdmin(user) && user.organizations.length > 1;

  const [relevancy, setRelevancy] = useState<number>(40);
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [activeTab, setActiveTab] = useState(searchParams.get('active_sidebar_menu') || SIDEBAR_MENU_LIST[0].active_menu);
  const [openSubredditDialog, setOpenSubredditDialog] = useState(false);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  console.log("###_debug_hoverActive ", hoverActive);

  const handleRelevancyChange = (_event: Event, newValue: number | number[]) => {
    setRelevancy(newValue as number)
  }

  useEffect(() => {
    const params = new URLSearchParams(searchParams);
    params.set('active_sidebar_menu', activeTab);
    router.push(`${pathname}?${params.toString()}`, { scroll: false });
  }, [activeTab, pathname, router, searchParams]);

  const handleMenuClick = (menu: string) => {
    const value = SIDEBAR_MENU_LIST.find(item => item?.name === menu)?.active_menu ?? SIDEBAR_MENU_LIST[0].active_menu;
    setActiveTab(value);
  };

  const isMenuActive = (currentMenu: string) => {
    const current = SIDEBAR_MENU_LIST.find(item => item?.name === currentMenu)?.active_menu;
    const isActive = searchParams?.get('active_sidebar_menu') === current;
    return isActive;
  }

  const handleOpenDialog = () => {
    setOpenSubredditDialog(true);
  }

  const handleClosDialog = () => {
    setOpenSubredditDialog(false);
  }

  const handleAdd = (subreddit: string) => {
    console.log(subreddit);
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
          width: 255,
          height: "100vh",
          borderRight: "1px solid #e0e0e0",
          borderRadius: 0,
          bgcolor: "white",
          px: 1.5
        }}
      >
        <Box sx={{ p: 2, display: "flex", flexDirection: "column", height: "100%" }}>
          {/* Avatar */}
          <Box sx={{ display: "flex", justifyContent: "flex-start", my: 5 }}>
            <Avatar sx={{ width: 48, height: 48, bgcolor: "#f3f4f6", color: "#111827" }}>A</Avatar>
          </Box>

          {/* Inbox */}
          <ListItem
            component={Box}
            sx={{
              borderRadius: 1.5,
              mb: 3,
              p: "8px 16px",
              bgcolor: isMenuActive('inbox') ? "#f9fafb" : "",
              boxShadow: isMenuActive('inbox') ? "rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0.05) 0px 1px 2px 0px" : "",
              "&:hover": {
                backgroundColor: "#F0F5FF"
              },
            }}
            onClick={() => handleMenuClick('inbox')}
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

          <List sx={{ p: 0, mb: "auto", display: "flex", flexDirection: "column", alignItems: "center", gap: 1.5 }}>
            {SUBREDDIT_LIST.map((ele, index) => (
              <ListItem
                key={index}
                disablePadding
                sx={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  width: "100%",
                  borderRadius: 1.5,
                  bgcolor: isMenuActive(ele.name) ? "#f9fafb" : "",
                  boxShadow: isMenuActive(ele.name) ? "rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0.05) 0px 1px 2px 0px" : "",
                  "&:hover": {
                    backgroundColor: "#F0F5FF"
                  },
                  "&:hover .hover-icon": {
                    opacity: 1,
                  },
                }}
                onClick={() => handleMenuClick(ele.name)}
              >
                <ListItemButton
                  sx={{
                    flexGrow: 1,
                    "&:hover": {
                      backgroundColor: "transparent"
                    },
                  }}
                >
                  <ListItemText primary={`${ele.prefix}${ele.name}`} />
                </ListItemButton>

                <Box sx={{ display: "flex", alignItems: "center", gap: 5 }}>
                  <Badge
                    badgeContent={ele.badge_count}
                    color="warning"
                    sx={{
                      "& .MuiBadge-badge": {
                        bgcolor: "#FF9800",
                        color: "white",
                        fontWeight: "bold",
                      },
                    }}
                  />
                  <IconButton
                    edge="end"
                    className="hover-icon"
                    disableRipple
                    sx={{
                      opacity: 0,
                      transition: "opacity 0.2s ease",
                    }}
                    onClick={handleClick}
                    size="small"
                    aria-controls={open ? 'account-menu' : undefined}
                    aria-haspopup="true"
                    aria-expanded={open ? 'true' : undefined}
                  >
                    <MoreVertIcon fontSize="small" />
                  </IconButton>
                  <Menu
                    anchorEl={anchorEl}
                    id="account-menu"
                    open={open}
                    onClose={handleClose}
                    onClick={handleClose}
                    slotProps={{
                      paper: {
                        elevation: 0,
                        sx: {
                          overflow: 'visible',
                          filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                          mt: 1.5,
                          '& .MuiAvatar-root': {
                            width: 32,
                            height: 32,
                            ml: -0.5,
                            mr: 1,
                          },
                          '&::before': {
                            content: '""',
                            display: 'block',
                            position: 'absolute',
                            top: 0,
                            right: 14,
                            width: 10,
                            height: 10,
                            bgcolor: 'background.paper',
                            transform: 'translateY(-50%) rotate(45deg)',
                            zIndex: 0,
                          },
                        },
                      },
                    }}
                    transformOrigin={{ horizontal: 'right', vertical: 'top' }}
                    anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
                  >
                    <MenuItem onClick={handleClose}>
                      <DeleteOutline /> {`Remove`}
                    </MenuItem>
                  </Menu>
                </Box>
              </ListItem>
            ))}
          </List>

          {/* Bottom Actions */}
          <Box sx={{ mt: 2 }}>
            <Divider sx={{ mb: 2 }} />
            <List sx={{ p: 0 }}>
              <ListItem onClick={handleOpenDialog} disablePadding sx={{ "&:hover": { backgroundColor: "#F0F5FF" } }}>
                <ListItemButton sx={{ borderRadius: 1.5 }}>
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

              <ListItem disablePadding sx={{ "&:hover": { backgroundColor: "#F0F5FF" } }}>
                <ListItemButton sx={{ borderRadius: 1.5 }}>
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

          {/*Add Subreddit Dialog */}
          <AddSubredditDialog open={openSubredditDialog} onClose={handleClosDialog} onAdd={handleAdd} />
        </Box>
      </Paper>
    </>
  )
}

export default NavBar
