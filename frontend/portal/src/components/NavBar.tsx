import { useCallback, useEffect, useMemo, useState, FC, MouseEvent } from 'react'
import { Fade, Menu, MenuItem, Skeleton, Tooltip } from '@mui/material'
import { ChevronDown, LogOut } from 'lucide-react'
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
  Divider,
  IconButton,
  Badge,
  Paper,
} from "@mui/material"
import {
  Mail as MailIcon,
  MoreVert as MoreVertIcon,
  Settings as SettingsIcon,
  RestartAlt as RestartAltIcon,
  DeleteOutline,
} from "@mui/icons-material"
import { usePathname, useRouter } from 'next/navigation';
// import AddSubredditDialog from './AddSubredditDialog';
import Link from 'next/link'
import { useClientsContext } from '@doota/ui-core/context/ClientContext'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { setError, setLoading, setSubredditList, SourceTyeps } from '../../store/Source/sourceSlice'
import { RootState } from '../../store/store'
import { isActivePath } from '../utils/url'
import { useDebounce } from '@doota/ui-core/hooks/useDebounce';
import { setRelevancyScore, setSubReddit } from '../../store/Params/ParamsSlice'
import { useRedditIntegrationStatus } from './Leads/Tabs/useRedditIntegrationStatus'
import { LeadTabStatus, setActiveTab, setCompletedList, setDiscardedTabList, setIsLoading, setNewTabList, setSelectedLeadData } from '../../store/Lead/leadSlice'
import { LeadStatus } from '@doota/pb/doota/core/v1/core_pb'
import { GetLeadsResponse } from '@doota/pb/doota/portal/v1/portal_pb'

export const LoadigSkeletons = ({ count, height }: { count: number, height: number | string }) => (
  [...Array(count)].map((_, i) => (
    <Skeleton key={i} variant="rounded" width="100%" height={height} />
  ))
);

const NavBar: FC = () => {
  const { user, logout } = useAuth()
  const { portalClient } = useClientsContext()
  const [currentOrg, setCurrentOrganization] = useOrganization()
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const open = Boolean(anchorEl)
  const dispatch = useAppDispatch();
  const router = useRouter();
  const pathname = usePathname();
  // const [openSubredditDialog, setOpenSubredditDialog] = useState(false);
  const [anchorEl1, setAnchorEl1] = useState<null | HTMLElement>(null);
  const open1 = Boolean(anchorEl1);
  const { subredditList, loading } = useAppSelector((state: RootState) => state.source);
  const { relevancyScore, subReddit } = useAppSelector((state: RootState) => state.parems);
  const { newTabList, isLoading } = useAppSelector((state: RootState) => state.lead);
  const [relevancy_score, setRelevancy_Score] = useState<number>(relevancyScore);
  const { isConnected, loading: isLoadingRedditIntegrationStatus } = useRedditIntegrationStatus();
  const canChangeOrg = user && isPlatformAdmin(user) && user.organizations.length > 1;
  const DEFAULT_RESET_RELEVENCY_SCORE = 90;

  const onChangeCommitted = useCallback((key: string, value: number | string) => {
    if (key === 'relevancy_score') {
      dispatch(setRelevancyScore(value as number));
    }
    if (key === 'currentActiveSubRedditId') {
      dispatch(setSubReddit(value as string));
    }
  }, [dispatch]);

  const debouncedOnChangeCommitted = useDebounce(onChangeCommitted, 300);

  const handleOpenOrg = (event: MouseEvent<HTMLDivElement>) => {
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

  const handleClick = (event: MouseEvent<HTMLElement>) => {
    setAnchorEl1(event.currentTarget);
  };

  const handleClose1 = () => {
    setAnchorEl1(null);
  };

  const handleRelevancyChange = (_event: Event, newValue: number | number[]) => {
    setRelevancy_Score(newValue as number);
    debouncedOnChangeCommitted('relevancy_score', newValue as number);
  }

  const handleSubRedditsClick = (data: SourceTyeps) => {
    debouncedOnChangeCommitted('currentActiveSubRedditId', data.id);
  };

  const handleMenuClick = () => {
    router.push('/dashboard/leads');
  };

  const isMenuActive = (id: string) => {
    const isActive = subReddit === id;
    return isActive;
  }

  // const handleClosDialog = () => {
  //   setOpenSubredditDialog(false);
  // }

  // const handleAdd = (subreddit: string) => {
  //   console.log(subreddit);
  // }

  useEffect(() => {
    const getAllSubReddits = async () => {
      dispatch(setLoading(true));
      dispatch(setSubReddit(""));

      try {
        const result = await portalClient.getSources({});
        const listOfSubReddits = result?.sources ?? [];
        dispatch(setSubredditList(listOfSubReddits));
      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong";
        toast.error(message);
        dispatch(setError(message));
      } finally {
        dispatch(setLoading(false));
      }
    };

    if (isConnected === true && isLoadingRedditIntegrationStatus === false) {
      getAllSubReddits();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConnected, isLoadingRedditIntegrationStatus]);

  const handleResetFilters = () => {
    // Reset any other filters here
    debouncedOnChangeCommitted('relevancy_score', DEFAULT_RESET_RELEVENCY_SCORE);
    setRelevancy_Score(DEFAULT_RESET_RELEVENCY_SCORE);
    debouncedOnChangeCommitted('currentActiveSubRedditId', "");
  }

  const isleads = isActivePath('/dashboard/leads', pathname);

  const subRedditsLoaded = useMemo(() => subredditList.length > 0 && !loading, [subredditList, loading]);

  useEffect(() => {
    if (!subRedditsLoaded) return; // Ensure we only proceed after subreddits are loaded

    const getAllRelevantLeads = async () => {
      dispatch(setIsLoading(true));
      dispatch(setActiveTab(LeadTabStatus.NEW));
      dispatch(setSelectedLeadData(null));

      try {
        const result = await portalClient.getRelevantLeads({
          ...(relevancyScore && { relevancyScore }),
          ...(subReddit && { subReddit })
        });
        const allLeads = result.leads ?? [];
        dispatch(setNewTabList(allLeads));

        // After relevant leads are fetched, trigger completed and discarded leads
        const completedResult = await getAllLeadsByStatus(LeadStatus.COMPLETED);
        const discardedResult = await getAllLeadsByStatus(LeadStatus.NOT_RELEVANT);

        dispatch(setCompletedList(completedResult.leads ?? []));
        dispatch(setDiscardedTabList(discardedResult?.leads ?? []));

        // Set the first lead only once after all API calls are complete
        if (allLeads.length > 0) {
          dispatch(setSelectedLeadData(allLeads[0])); // Set the first lead as selected
        }

      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong";
        toast.error(message);
        dispatch(setError(message));
      } finally {
        dispatch(setIsLoading(false));
      }
    };

    const getAllLeadsByStatus = async (status: LeadStatus): Promise<GetLeadsResponse> => {
      try {
        const result = await portalClient.getLeadsByStatus({ status });
        return result;
      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong";
        throw new Error(message);
      }
    };

    getAllRelevantLeads();

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [relevancyScore, subReddit, subRedditsLoaded]);

  const newTabListLoaded = useMemo(() => newTabList.length > 0 && !isLoading, [newTabList, isLoading]);

  const countItem = useCallback((id: string) => {
    if (!newTabListLoaded) return 0;
    return id === "posts"
      ? newTabList.length
      : newTabList.reduce((count, ele) => ele.sourceId === id ? count + 1 : count, 0);
  }, [newTabListLoaded, newTabList]);

  return (<>
    <Paper
      elevation={0}
      sx={{
        width: 255,
        height: "100vh",
        borderRight: "1px solid #e0e0e0",
        borderRadius: 0,
        bgcolor: "rgb(17 24 39)",
      }}
    >
      <Box sx={{ py: 3, display: "flex", flexDirection: "column", height: "100%" }}>

        {/* Avatar */}
        <Box sx={{ px: 5, display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div className='text-black flex font-extrabold h-14 items-center w-full'>
            <div
              role='presentation'
              className={`flex justify-between w-full relative text-xl truncate h-full items-center`}
              onClick={canChangeOrg ? handleOpenOrg : undefined}
            >
              <Tooltip
                title='Select your organization'
                placement='right'
                arrow
                disableInteractive
                disableHoverListener={canChangeOrg ? false : true}
              >
                <div className='flex items-center gap-[4px] text-gray-100'>
                  <div className=''>{currentOrg?.name}</div>
                  {canChangeOrg && (
                    <div className={`inline-block`}>
                      <ChevronDown size={18} />
                    </div>
                  )}
                </div>
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
          <Tooltip title='Log Out' placement='right' arrow disableInteractive>
            <LogOut className='text-gray-100 cursor-pointer' size={14} onClick={handleLogout} />
          </Tooltip>
        </Box>
        <Divider sx={{ bgcolor: "#2d3748" }} />

        {/* Posts */}
        <List
          sx={{
            m: 2,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <ListItem
            disablePadding
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              width: "100%",
              borderRadius: 1.5,
              bgcolor: isleads ? "#1f2937" : "",
              boxShadow: isleads ? "rgba(0, 0, 0, 0.05) 0px 1px 2px 0px" : "",
              "&:hover": {
                backgroundColor: "#1f2937",
                "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                  color: isleads ? "white" : "#d1d5db",
                },
              },
              "& .MuiListItemSecondaryAction-root": {
                right: "26px",
              },
            }}
            onClick={() => handleMenuClick()}
            secondaryAction={
              <Badge
                badgeContent={countItem("posts")}
                color="warning"
                sx={{
                  "& .MuiBadge-badge": {
                    bgcolor: "#FF9800",
                    color: "white",
                    fontWeight: "bold",
                  },
                }}
                max={Infinity}
              />
            }
          >
            <ListItemButton
              sx={{
                flexGrow: 1,
                pl: 2,
                "&:hover": {
                  backgroundColor: "transparent",
                },
                // Apply white color when active
                "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                  color: isleads ? "white" : "#d1d5db",
                },
              }}
            >
              <ListItemIcon sx={{ minWidth: "auto", mr: 2 }}>
                <MailIcon />
              </ListItemIcon>
              <ListItemText primary={`Posts`} />
            </ListItemButton>
          </ListItem>
        </List>

        {/* Filters */}
        <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2, mx: 4.5 }}>
          <Typography variant="subtitle2" sx={{ color: "#9ca3af" }}>
            FILTERS
          </Typography>
          <Tooltip
            title="Reset all filters"
            placement="right"
            TransitionComponent={Fade}
            TransitionProps={{ timeout: 600 }}
            arrow
          >
            <IconButton
              size="small"
              onClick={handleResetFilters}
              aria-label="Reset filters"
              sx={{
                color: "#9ca3af",
                "&:hover": {
                  color: "primary",
                  backgroundColor: "rgba(63, 81, 181, 0.04)",
                },
              }}
            >
              <RestartAltIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>

        <Box sx={{ mb: 3, mx: 4.5 }}>
          <Box sx={{ display: "flex", justifyContent: "space-between", mb: 2 }}>
            <Typography variant="body2" sx={{ color: "#9ca3af" }}>Relevancy</Typography>
            <Typography variant="body2" sx={{ color: "#9ca3af" }}>
              {relevancy_score}%
            </Typography>
          </Box>
          <Slider
            value={relevancy_score}
            onChange={handleRelevancyChange}
            min={70}
            step={10}
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
        <Typography variant="subtitle2" sx={{ mb: 2, mx: 4.5, color: "#9ca3af" }}>
          FILTER BY SUBREDDIT
        </Typography>

        {(loading || isLoadingRedditIntegrationStatus) ?
          <Box sx={{ display: 'flex', px: 2, flexDirection: "column", alignItems: "center", height: "100%", width: "100%", gap: 2 }}>
            <LoadigSkeletons count={3} height={40} />
          </Box>
          :
          subredditList?.length > 0 ?
            <List sx={{ px: 2, mb: "auto", display: "flex", flexDirection: "column", alignItems: "center", gap: 1.5 }}>
              {subredditList.map((ele, index) => (
                <ListItem
                  key={index}
                  disablePadding
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    width: "100%",
                    borderRadius: 1.5,
                    px: 4,
                    bgcolor: isMenuActive(ele.id) ? "#1f2937" : "",
                    boxShadow: isMenuActive(ele.id) ? "rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0.05) 0px 1px 2px 0px" : "",
                    "&:hover": {
                      backgroundColor: "#1f2937",
                      "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                        color: isMenuActive(ele.id) ? "white" : "#d1d5db",
                      },
                    },
                    "& .hover-icon": {
                      color: "#d1d5db"
                    },
                    "&:hover .hover-icon": {
                      opacity: 1,
                      color: "white"
                    },
                    "& .MuiListItemSecondaryAction-root": {
                      right: "8px",
                    },
                  }}
                  onClick={() => handleSubRedditsClick(ele)}
                  secondaryAction={
                    <Box sx={{ display: "flex", alignItems: "center", gap: 5, width: "auto" }}>
                      <Badge
                        badgeContent={countItem(ele.id)}
                        color="warning"
                        sx={{
                          "& .MuiBadge-badge": {
                            bgcolor: "#FF9800",
                            color: "white",
                            fontWeight: "bold",
                          },
                        }}
                        max={Infinity}
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
                        aria-controls={open1 ? 'account-menu' : undefined}
                        aria-haspopup="true"
                        aria-expanded={open1 ? 'true' : undefined}
                      >
                        <MoreVertIcon fontSize="small" />
                      </IconButton>
                      <Menu
                        anchorEl={anchorEl}
                        id="account-menu"
                        open={open1}
                        onClose={handleClose1}
                        onClick={handleClose1}
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
                        <MenuItem onClick={handleClose1}>
                          <DeleteOutline /> {`Remove`}
                        </MenuItem>
                      </Menu>
                    </Box>
                  }
                >
                  <ListItemButton
                    sx={{
                      flexGrow: 1,
                      px: 0,
                      paddingRight: 0,
                      "&:hover": {
                        backgroundColor: "transparent"
                      },
                      "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                        color: isMenuActive(ele.id) ? "white" : "#d1d5db",
                      },
                    }}
                  >
                    <ListItemText primary={`${ele.name}`} />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
            :
            <Box sx={{ display: 'flex', flexDirection: "column", alignItems: "center", height: "100%", width: "100%", mt: 3 }}>
              <Typography variant="subtitle2" color="text.secondary">
                {`No subreddits avalable.`}
              </Typography>
            </Box>
        }

        {/* Bottom Actions */}
        <Box sx={{ mt: 2 }}>
          <Divider sx={{ mb: 2, bgcolor: "#2d3748" }} />
          <List disablePadding sx={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 1.5, mx: 2 }}>
            {/* <ListItem onClick={handleOpenDialog} disablePadding sx={{ "&:hover": { backgroundColor: "#F0F5FF" } }}>
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
              </ListItem> */}

            {/* Navigation */}
            {user && isAdmin(user) && (<>
              <ListItem
                component={Link}
                href={routes.app.settings.account}
                disablePadding
                sx={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  width: "100%",
                  borderRadius: 1.5,
                  "&:hover": {
                    backgroundColor: "#1f2937",
                    "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                      color: "#d1d5db",
                    },
                  }
                }}
              >
                <ListItemButton
                  sx={{
                    flexGrow: 1,
                    pl: 2,
                    "&:hover": {
                      backgroundColor: "transparent",
                    },
                    "& .MuiListItemIcon-root, & .MuiListItemText-primary": {
                      color: "#d1d5db",
                    },
                  }}
                >
                  <ListItemIcon sx={{ minWidth: "auto", mr: 2 }}>
                    <SettingsIcon />
                  </ListItemIcon>
                  <ListItemText primary={`Settings`} />
                </ListItemButton>
              </ListItem>
            </>)}
          </List>
        </Box>

        {/*Add Subreddit Dialog */}
        {/* <AddSubredditDialog open={openSubredditDialog} onClose={handleClosDialog} onAdd={handleAdd} /> */}
      </Box>
    </Paper>
  </>);
}

export default NavBar
