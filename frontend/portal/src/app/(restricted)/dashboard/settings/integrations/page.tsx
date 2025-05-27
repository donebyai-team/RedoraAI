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
import { IntegrationType, Integration } from '@doota/pb/doota/portal/v1/portal_pb'
import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner'
import { Button } from '../../../../../atoms/Button'
import { portalClient } from '../../../../../services/grpc'
import { buildAppUrl } from '../../../../routes'
import { routes } from '@doota/ui-core/routing'
import { isAdmin, isPlatformAdmin } from '@doota/ui-core/helper/role'
import { Box } from '@mui/system'
import { AppBar, Toolbar, Typography, Card, CardContent, Slider, Switch, styled } from '@mui/material'
import {
  Reddit as RedditIcon,
} from "@mui/icons-material"
import toast from 'react-hot-toast'

const StyledSlider = styled(Slider)(() => ({
  color: '#111827', // Dark color for the track
  height: 8,
  '& .MuiSlider-track': {
    border: 'none',
    backgroundColor: '#111827',
  },
  '& .MuiSlider-thumb': {
    height: 24,
    width: 24,
    backgroundColor: '#fff',
    border: '2px solid currentColor',
    '&:focus, &:hover, &.Mui-active, &.Mui-focusVisible': {
      boxShadow: '0 0 0 8px rgba(0, 0, 0, 0.1)',
    },
  },
  '& .MuiSlider-rail': {
    color: '#d1d5db',
    opacity: 1,
  },
}));

const SaveButton = styled(Button)(() => ({
  background: 'linear-gradient(90deg, #800080 0%, #9333ea 100%)',
  color: 'white',
  fontWeight: 'bold',
  textTransform: 'none',
  padding: '10px 24px',
  marginTop: '12px',
  borderRadius: '8px',
  '&:hover': {
    background: 'linear-gradient(90deg, #6b016b 0%, #7929c4 100%)',
  },
}));

const CustomSwitch = styled(Switch)(() => ({
  width: 42,
  height: 26,
  padding: 0,
  '& .MuiSwitch-switchBase': {
    padding: 0,
    margin: 2,
    transitionDuration: '300ms',
    '&.Mui-checked': {
      transform: 'translateX(16px)',
      color: '#fff',
      '& + .MuiSwitch-track': {
        backgroundColor: '#111827',
        opacity: 1,
        border: 0,
      },
    },
  },
  '& .MuiSwitch-thumb': {
    boxSizing: 'border-box',
    width: 22,
    height: 22,
  },
  '& .MuiSwitch-track': {
    borderRadius: 26 / 2,
    backgroundColor: '#a1a1aa',
    opacity: 1,
  },
}));

const defaultRelevancyScoreForComment = 90;
const defaultStatusForComment = false;

export default function Page() {
  const user = useAuthUser()
  const { setUser, setOrganization, getOrganization } = useAuth()

  const [loading, setLoading] = useState(true)
  const [integrations, setIntegrations] = useState<Integration[]>([])

  const org = getOrganization();

  const defaultRelevancyScore = org?.featureFlags?.Comment?.relevancyScore ?? defaultRelevancyScoreForComment;
  const defaultAutoComment = org?.featureFlags?.Comment?.enabled ?? defaultStatusForComment;

  const [relevancyScore, setRelevancyScore] = useState(defaultRelevancyScore)
  const [autoComment, setAutoComment] = useState(defaultAutoComment)

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

  const openOauthConsentScreen = (integrationType: IntegrationType) => {
    portalClient
      .oauthAuthorize({
        integrationType: integrationType,
        redirectUrl: buildAppUrl(routes.app.home),
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
      })
  }

  const handleScoreChange = (_event: Event, newValue: number | number[]) => {
    setRelevancyScore(newValue as number);
  };

  const handleSaveAutomation = async () => {
    try {
      const result = await portalClient.updateAutomationSettings({
        comment: { enabled: autoComment, relevancyScore }
      });

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

      toast.success("Automation settings updated successfully!")
    } catch (err) {
      if (err instanceof Error) {
        const message = err.message || "Failed to update automation settings";
        toast.error(message);
      } else {
        console.error("Unexpected error:", err)
      }

    }
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
              </TableRow>
            </TableHead>
            <TableBody>
              {integrations.map((row, index) => (
                <TableRow key={index}>
                  <TableCell>Reddit</TableCell>
                  <TableCell>{row.details.value?.userName}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        <Card sx={{ p: 4, mt: 5 }} component={Paper}>
          <CardContent>
            <Box display="flex" alignItems="center" gap={1} mb={2}>
              <Typography variant="h4" component="div" fontWeight="bold">
                Automated Comments Settings
              </Typography>
            </Box>

            <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
              Configure your automation preferences.
            </Typography>

            <Box mb={5}>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="body1" fontWeight="medium">
                  Minimum Relevancy Score: {relevancyScore}%
                </Typography>
              </Box>

              <StyledSlider
                value={relevancyScore}
                onChange={handleScoreChange}
                min={80}
                max={100}
                step={5}
                aria-label="Relevancy Score"
              />

              <Typography variant="body2" color="text.secondary" sx={{ mt: 1.5 }}>
                RedoraAI will automatically post comments on posts ≥ Min Relevancy Score
              </Typography>
            </Box>

            <Box display="flex" alignItems="center" py={2} mb={4}>
              <CustomSwitch
                checked={autoComment}
                onChange={(e) => setAutoComment(e.target.checked)}
              />
              <Typography variant="body1" fontWeight="medium" ml={2.5} display={"flex"}>
                {'Automated Comments'}
                <Typography variant="body1" fontWeight="medium" ml={1} sx={{ color: autoComment ? "green" : "red" }} >
                  {`${autoComment ? "On" : "Off"}`}
                </Typography>
              </Typography>
            </Box>

            <SaveButton onClick={handleSaveAutomation} variant="contained" size="large">
              Save Automation Settings
            </SaveButton>
          </CardContent>
        </Card>

        <Card sx={{ p: 4, mt: 5 }} component={Paper}>
          <CardContent>
            <Box display="flex" alignItems="center" gap={1} mb={2}>
              <Typography variant="h4" component="div" fontWeight="bold">
                Automated DM Settings
              </Typography>
            </Box>

            <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
              Contact us via chat to configure automated DM.
            </Typography>
          </CardContent>
        </Card>

      </Box>
    </Box>
  </>);
}
