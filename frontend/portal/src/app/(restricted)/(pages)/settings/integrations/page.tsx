import { RedditAccountsList } from "@/components/reddit-accounts/RedditAccountsList";

export default function Page() {
  return (<RedditAccountsList />);
}


// 'use client'

// import { useEffect, useState } from 'react'
// import Table from '@mui/material/Table'
// import TableBody from '@mui/material/TableBody'
// import TableCell from '@mui/material/TableCell'
// import TableContainer from '@mui/material/TableContainer'
// import TableHead from '@mui/material/TableHead'
// import TableRow from '@mui/material/TableRow'
// import Paper from '@mui/material/Paper'
// import { useAuth, useAuthUser } from '@doota/ui-core/hooks/useAuth'
// import { IntegrationType, Integration, IntegrationState, Organization } from '@doota/pb/doota/portal/v1/portal_pb'
// import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner'
// import { Button } from '../../../../../atoms/Button'
// import { portalClient } from '../../../../../services/grpc'
// import { buildAppUrl } from '../../../../routes'
// import { routes } from '@doota/ui-core/routing'
// import { isAdmin, isPlatformAdmin } from '@doota/ui-core/helper/role'
// import { Box } from '@mui/system'
// import { AppBar, Toolbar, Typography } from '@mui/material'
// import {
//   Reddit as RedditIcon,
// } from "@mui/icons-material"
// import toast from 'react-hot-toast'
// import { getConnectError } from '@/utils/error'

// export default function Page() {
//   const user = useAuthUser()
//   const { setUser, setOrganization } = useAuth()

//   const [loading, setLoading] = useState(true)
//   const [integrations, setIntegrations] = useState<Integration[]>([])

//   useEffect(() => {
//     portalClient.getIntegrations({})
//       .then((res) => {
//         setIntegrations(res.integrations);
//       })
//       .catch((err) => {
//         console.error("Error fetching integrations:", err);
//       })
//       .finally(() => {
//         setLoading(false);
//       });
//   }, []);

//   const getIntegrationByType = (
//     integrations: Integration[],
//     integrationType: IntegrationType
//   ): Integration | undefined => {
//     return integrations.find((integration) => integration.type === integrationType && integration.status == IntegrationState.ACTIVE);
//   };

//   const handleDisconnectReddit = (id: string) => {
//     // Optimistically update the state to AUTH_REMOVED
//     setIntegrations((prev) =>
//       prev.map((i) =>
//         i.id === id ? { ...i, status: IntegrationState.AUTH_REVOKED } : i
//       )
//     );

//     // Send API call async
//     portalClient.revokeIntegration({ id: id })
//       .then(() => {
//         // After successful revoke, check current state
//         setIntegrations((prev) => {
//           const updated = prev.map((i) =>
//             i.id === id ? { ...i, status: IntegrationState.AUTH_REVOKED } : i
//           );

//           const hasActiveReddit = updated.some(
//             (i) => i.type === IntegrationType.REDDIT && i.status === IntegrationState.ACTIVE
//           );

//           if (!hasActiveReddit) {
//             console.log("No active reddit account, disabling automated comments");
//             handleSaveAutomation({ comment: { enabled: false } });
//           }

//           return updated;
//         });
//       })
//       .catch((err) => {
//         toast.error(getConnectError(err));
//       });
//   };


//   // TODO: This is duplicate in automation/page as well, merge it into common
//   const handleSaveAutomation = async (req: any) => {
//     try {
//       console.log("Updating autmation", req);
//       const result = await portalClient.updateAutomationSettings(req);

//       if (isPlatformAdmin(user)) {
//         setOrganization(result);
//       }

//       setUser(prev => {
//         if (!prev) return prev
//         const updatedOrganizations = prev.organizations.map(org =>
//           org.id === result.id ? result : org
//         )
//         return { ...prev, organizations: updatedOrganizations }
//       })
//     } catch (err) {
//       toast.error(getConnectError(err));
//     }
//   }

//   const openOauthConsentScreen = (integrationType: IntegrationType) => {
//     portalClient
//       .oauthAuthorize({
//         integrationType: integrationType,
//         redirectUrl: buildAppUrl(routes.new.dashboard),
//       })
//       .then(oAuthAuthorizeResp => {
//         window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
//       })
//   }

//   if (loading) {
//     return <FallbackSpinner />
//   }

//   return (<>
//     <Box component="main" sx={{ flexGrow: 1, p: 0, display: "flex", flexDirection: "column" }}>
//       <AppBar position="static" color="inherit" elevation={0} sx={{ borderBottom: "1px solid #e0e0e0", height: 61 }}>
//         <Toolbar>
//           <Box sx={{ flexGrow: 1 }} />
//           <Typography variant="body2" color="text.secondary" sx={{ mr: 2 }}>
//             {user && user.email}
//           </Typography>
//           {user && isAdmin(user) && (<>
//             <Button
//               variant="contained"
//               startIcon={<RedditIcon />}
//               sx={{
//                 bgcolor: "#ff4500",
//                 "&:hover": {
//                   bgcolor: "#e03d00",
//                 },
//                 gap: 0
//               }}
//               onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
//             >
//               Connect Reddit
//             </Button>
//           </>)}
//         </Toolbar>
//       </AppBar>

//       <Box sx={{ p: 3, flexGrow: 1 }}>
//         <Paper
//           variant="outlined"
//           sx={{
//             p: 2,
//             mb: 2,
//             borderLeft: "4px solid #ff4500",
//             backgroundColor: "#fff8f6",
//           }}
//         >
//           <Typography variant="body2" sx={{ color: "#4d2c19" }}>
//             <strong>Note:</strong> You can connect multiple Reddit accounts. We will automatically rotate between them when sending comments.
//           </Typography>
//         </Paper>

//         <TableContainer component={Paper} elevation={0} variant="outlined">
//           <Table sx={{ minWidth: 650 }}>
//             <TableHead>
//               <TableRow>
//                 <TableCell sx={{ fontWeight: "medium" }}>Provider</TableCell>
//                 <TableCell sx={{ fontWeight: "medium" }}>Username</TableCell>
//                 <TableCell sx={{ fontWeight: "medium" }}>State</TableCell>
//                 <TableCell /> {/* Empty header for action column */}
//               </TableRow>
//             </TableHead>
//             <TableBody>
//               {integrations
//                 .filter((i) => i.type === IntegrationType.REDDIT)
//                 .map((integration) => {
//                   const isActive = integration.status === IntegrationState.ACTIVE
//                   const isAuthRemoved = integration.status !== IntegrationState.ACTIVE
//                   const username = integration.details?.value?.userName ?? '—'
//                   const reason = integration.details?.value?.reason;

//                   return (
//                     <TableRow key={integration.id}>
//                       <TableCell>Reddit</TableCell>
//                       <TableCell>
//                         {username !== '—' ? (
//                           <div className="flex flex-col">
//                             <a
//                               href={`https://reddit.com/user/${username}`}
//                               target="_blank"
//                               rel="noopener noreferrer"
//                               className="text-blue-600 hover:underline"
//                             >
//                               {username}
//                             </a>
//                             {reason && (
//                               <span className="text-xs text-yellow-700 mt-1">
//                                 {reason}
//                               </span>
//                             )}
//                           </div>
//                         ) : (
//                           '—'
//                         )}
//                       </TableCell>
//                       <TableCell>
//                         {integration.status === IntegrationState.ACTIVE ? 'Active' : 'Revoked'}
//                       </TableCell>
//                       <TableCell align="right">
//                         {isActive ? (
//                           <Button
//                             color="error"
//                             variant="outlined"
//                             size="small"
//                             onClick={() => handleDisconnectReddit(integration.id)}
//                           >
//                             Disconnect
//                           </Button>
//                         ) : isAuthRemoved ? (
//                           <Button
//                             color="primary"
//                             variant="outlined"
//                             size="small"
//                             onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
//                           >
//                             Reconnect
//                           </Button>
//                         ) : null}
//                       </TableCell>
//                     </TableRow>
//                   )
//                 })}
//             </TableBody>

//           </Table>
//         </TableContainer>
//       </Box>
//     </Box>
//   </>);
// }

