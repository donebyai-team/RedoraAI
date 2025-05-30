'use client'

import { useEffect, useState } from 'react'
import Paper from '@mui/material/Paper'
import { useAuth, useAuthUser } from '@doota/ui-core/hooks/useAuth'
import { IntegrationType, Integration, IntegrationState } from '@doota/pb/doota/portal/v1/portal_pb'
import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner'
import { Button } from '../../../../../atoms/Button'
import { portalClient } from '../../../../../services/grpc'
import { isPlatformAdmin } from '@doota/ui-core/helper/role'
import { Box } from '@mui/system'
import { Typography, Card, CardContent, Slider, Switch, styled } from '@mui/material'
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
    const [isConnecting, setIsConnecting] = useState(false)

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

    const getIntegrationByType = (
        integrations: Integration[],
        integrationType: IntegrationType
    ): Integration | undefined => {
        return integrations.find((integration) => integration.type === integrationType && integration.status == IntegrationState.ACTIVE);
    };

    const handleConnectReddit = async () => {
        try {
            setIsConnecting(true)
            let popup: Window | null = null;
            popup = window.open('', '_blank', "width=600,height=800");
            if (!popup) {
                toast.error('Popup was blocked. Please allow popups in your browser.');
                return;
            }
            // Inject temporary loading UI
            popup.document.write(`
                <html>
                    <head><title>Connecting...</title></head>
                    <body style="display:flex;justify-content:center;align-items:center;height:100vh;font-family:sans-serif;">
                <div>
                    <p>Connecting to Reddit Chat...</p>
                </div>
                </body>
                </html>
            `);
            popup.document.close();
            const abortController = new AbortController();
            const response = portalClient.connectReddit({}, { signal: abortController.signal });

            let streamClosed = false;

            // Set interval to check if popup closed manually
            const popupCheckInterval = setInterval(() => {
                if (popup && popup.closed && !streamClosed) {
                    // User closed popup before connection finished
                    console.log("Popup closed manually, canceling stream...");
                    // Cancel the stream here if possible
                    // Note: Depending on your gRPC lib, this might be a cancel() method or similar
                    // For example: response.cancel();

                    setIsConnecting(false);
                    clearInterval(popupCheckInterval);
                    streamClosed = true;
                    abortController.abort(); // ⛔ cancels the stream
                }
            }, 500); // check every 500ms

            for await (const msg of response) {
                if (msg.url) {
                    // Open the Reddit login page in a popup
                    popup.location.href = msg.url; // Redirect once URL is available
                }
            }

            // Stream finished normally
            streamClosed = true;
            clearInterval(popupCheckInterval);
            // Stream has ended successfully
            if (popup && !popup.closed) {
                popup.close();
            }

            await handleSaveAutomation({ dm: { enabled: true } })

            // Refresh integrations to reflect the newly connected status
            const res = await portalClient.getIntegrations({})
            setIntegrations(res.integrations)

            // Optionally show a success message or refetch integration status
            toast.success("Reddit connected successfully");
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        } finally {
            setIsConnecting(false)
        }
    };

    const handleDisconnectReddit = (id: string) => {
        // immediately remove
        setIntegrations((prev) =>
            prev.filter((i) => i.id !== id)
        )
        // send api call async
        portalClient.revokeIntegration({ id: id })
            .then(() => {
                handleSaveAutomation({ dm: { enabled: false } })
                console.log("successfully revoked")
            })
            .catch((err) => {
                console.error("Error disconnecting integrations:", err);
            })
    }

    const handleScoreChange = (_event: Event, newValue: number | number[]) => {
        setRelevancyScore(newValue as number);
    };

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

            <Box sx={{ p: 3, flexGrow: 1 }}>
                {/* DM automation settings */}
                <Card sx={{ p: 4, mt: 5 }} component={Paper}>
                    <CardContent>
                        <Box display="flex" alignItems="center" gap={1} mb={2}>
                            <Typography variant="h4" fontWeight="bold">
                                DM Automation Settings
                            </Typography>
                        </Box>

                        <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                            Please allow popup window to login into Reddit chat.
                        </Typography>

                        <Box display="flex" alignItems="center" gap={2}>
                            {getIntegrationByType(integrations, IntegrationType.REDDIT_DM_LOGIN) ? (
                                <>
                                    <Typography variant="body2" color="green" fontWeight="bold">
                                        Connected
                                    </Typography>
                                    <Button
                                        variant="outlined"
                                        color="error"
                                        size="small"
                                        onClick={() => handleDisconnectReddit(getIntegrationByType(integrations, IntegrationType.REDDIT_DM_LOGIN)?.id!)}
                                    >
                                        Disconnect
                                    </Button>
                                </>
                            ) : (
                                <SaveButton
                                    onClick={handleConnectReddit}
                                    variant="contained"
                                    size="large"
                                    disabled={isConnecting}
                                >
                                    {isConnecting ? 'Connecting...' : 'Connect Reddit DM'}
                                </SaveButton>
                            )}
                        </Box>
                    </CardContent>
                </Card>

                {/* Comment automation settings */}
                <Card sx={{ p: 4, mt: 5, mb: 10 }} component={Paper}>
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

                        <SaveButton
                            onClick={() => {
                                handleSaveAutomation({ comment: { enabled: autoComment, relevancyScore } })
                            }}
                            variant="contained" size="large">
                            Save Automation Settings
                        </SaveButton>
                    </CardContent>
                </Card>
            </Box>
        </Box>
    </>);
}
