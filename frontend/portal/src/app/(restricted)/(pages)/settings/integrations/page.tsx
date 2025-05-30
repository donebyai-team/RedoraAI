'use client'

import { useEffect, useState } from 'react'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Paper from '@mui/material/Paper'
import { useAuth, useAuthUser } from '@doota/ui-core/hooks/useAuth'
import { IntegrationType, Integration, IntegrationState, Organization } from '@doota/pb/doota/portal/v1/portal_pb'
import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner'
import { Button } from '../../../../../atoms/Button'
import { portalClient } from '../../../../../services/grpc'
import { buildAppUrl } from '../../../../routes'
import { routes } from '@doota/ui-core/routing'
import { isAdmin, isPlatformAdmin } from '@doota/ui-core/helper/role'
import { Box } from '@mui/system'
import { AppBar, Toolbar, Typography } from '@mui/material'
import {
  Reddit as RedditIcon,
} from "@mui/icons-material"
import toast from 'react-hot-toast'

export default function Page() {
  const user = useAuthUser()
  const { setUser, setOrganization } = useAuth()

  const [loading, setLoading] = useState(true)
  const [integrations, setIntegrations] = useState<Integration[]>([])

  useEffect(() => {
    portalClient.getIntegrations({})
      .then((res) => {
        setIntegrations(res.integrations);
      })
      .catch((err) => {
        console.error("Error fetching integrations:", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  const getIntegrationByType = (
    integrations: Integration[],
    integrationType: IntegrationType
  ): Integration | undefined => {
    return integrations.find((integration) => integration.type === integrationType && integration.status == IntegrationState.ACTIVE);
  };

  const handleDisconnectReddit = (id: string) => {
    // immediately remove
    setIntegrations((prev) =>
      prev.filter((i) => i.id !== id)
    )
    // send api call async
    portalClient.revokeIntegration({ id: id })
      .then(() => {
        handleSaveAutomation({ Comment: { enable: false }, dm: { enable: false } })
        console.log("successfully revoked")
      })
      .catch((err) => {
        console.error("Error disconnecting integrations:", err);
      })
  }

  // TODO: This is duplicate in automation/page as well, merge it into common
  const handleSaveAutomation = async (req: any) => {
    try {
      const result = await portalClient.updateAutomationSettings(req);

      if (isPlatformAdmin(user)) {
        setOrganization(result);
      }

      setUser(prev => {
        if (!prev) return prev
        const updatedOrganizations = prev.organizations.map(org =>
          org.id === result.id ? result : org
        )
        return { ...prev, organizations: updatedOrganizations }
      })
    } catch (err) {
      if (err instanceof Error) {
        const message = err.message || "Failed to update automation settings";
        toast.error(message);
      } else {
        console.error("Unexpected error:", err)
      }

    }
  }

  const openOauthConsentScreen = (integrationType: IntegrationType) => {
    portalClient
      .oauthAuthorize({
        integrationType: integrationType,
        redirectUrl: buildAppUrl(routes.new.dashboard),
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
      })
  }

  if (loading) {
    return <FallbackSpinner />
  }

  return (<>
    <Box component="main" sx={{ flexGrow: 1, p: 0, display: "flex", flexDirection: "column" }}>
      <AppBar position="static" color="inherit" elevation={0} sx={{ borderBottom: "1px solid #e0e0e0", height: 61 }}>
        <Toolbar>
          <Box sx={{ flexGrow: 1 }} />
          <Typography variant="body2" color="text.secondary" sx={{ mr: 2 }}>
            {user && user.email}
          </Typography>
          {user && isAdmin(user) && (<>
            <Button
              variant="contained"
              startIcon={<RedditIcon />}
              sx={{
                bgcolor: "#ff4500",
                "&:hover": {
                  bgcolor: "#e03d00",
                },
                gap: 0
              }}
              onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
            >
              Connect Reddit
            </Button>
          </>)}
        </Toolbar>
      </AppBar>

      <Box sx={{ p: 3, flexGrow: 1 }}>
        <TableContainer component={Paper} elevation={0} variant="outlined">
          <Table sx={{ minWidth: 650 }}>
            <TableHead>
              <TableRow>
                <TableCell sx={{ fontWeight: "medium" }}>Provider</TableCell>
                <TableCell sx={{ fontWeight: "medium" }}>Username</TableCell>
                <TableCell /> {/* Empty header for action column */}
              </TableRow>
            </TableHead>
            <TableBody>
              {(() => {
                const redditIntegration = getIntegrationByType(integrations, IntegrationType.REDDIT);
                if (redditIntegration) {
                  return (
                    <TableRow key="reddit">
                      <TableCell>Reddit</TableCell>
                      <TableCell>{redditIntegration.details?.value?.userName || '—'}</TableCell>
                      <TableCell align="right">
                        <Button
                          color="error"
                          variant="outlined"
                          size="small"
                          onClick={() => handleDisconnectReddit(redditIntegration.id)}
                        >
                          Disconnect
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                }
                return null;
              })()}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    </Box>
  </>);
}

