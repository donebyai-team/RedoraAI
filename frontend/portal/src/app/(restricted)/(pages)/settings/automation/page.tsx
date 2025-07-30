'use client';

import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { useAuth, useAuthUser } from '@doota/ui-core/hooks/useAuth';
import { IntegrationType, Integration, IntegrationState, NotificationFrequency } from '@doota/pb/doota/portal/v1/portal_pb';
import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner';
import { Button } from '../../../../../atoms/Button';
import { portalClient } from '../../../../../services/grpc';
import { isPlatformAdmin } from '@doota/ui-core/helper/role';
import { Box } from '@mui/system';
import { Typography, Card, CardContent, Slider, Switch, styled, Tabs, Tab, FormControlLabel, RadioGroup, Radio, Dialog, DialogTitle, DialogContent, DialogActions, TableCell, TableRow, TableBody, TableContainer, Table, TableHead, TextField, Link, List, ListItem, ListItemText } from '@mui/material';
import toast from 'react-hot-toast';
import { useAppSelector } from '@/store/hooks';
import { SubscriptionStatus } from '@doota/pb/doota/core/v1/core_pb';
import { getConnectError } from '@/utils/error';
import { LoadingButton } from '@mui/lab';

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
    const user = useAuthUser();
    const { setUser, setOrganization, getOrganization } = useAuth();

    const [loading, setLoading] = useState(true);
    const [integrations, setIntegrations] = useState<Integration[]>([]);
    const [isConnecting, setIsConnecting] = useState(false);
    const [currentTab, setCurrentTab] = useState(0); // 0 for Automation, 1 for Notification
    const project = useAppSelector((state) => state.stepper.project);

    const org = getOrganization();

    const hasPlanExpired = (org && org?.featureFlags?.subscription?.status === SubscriptionStatus.EXPIRED) ?? false;
    const defaultRelevancyScore = org?.featureFlags?.Comment?.relevancyScore ?? defaultRelevancyScoreForComment;
    const defaultAutoComment = org?.featureFlags?.Comment?.enabled ?? defaultStatusForComment;
    const defaultPostFrequency = org?.featureFlags?.notificationSettings?.relevantPostFrequency ?? NotificationFrequency.DAILY;
    const maxDMPerDay = org?.featureFlags?.DM?.maxPerDay || 0;
    const maxCommentPerDayLimit = org?.featureFlags?.subscription?.comments?.perDay || 0;
    const maxCommentPerDay = org?.featureFlags?.Comment?.maxPerDay || 5;
    const [maxCommentsInput, setMaxCommentsInput] = useState(maxCommentPerDay.toString());

    const [relevancyScore, setRelevancyScore] = useState(defaultRelevancyScore);
    const [autoComment, setAutoComment] = useState(defaultAutoComment);

    // Notification settings states
    // Initialize with a default or fetched value for the actual project's setting
    const [emailFrequency, setEmailFrequency] = useState<NotificationFrequency>(defaultPostFrequency); // Default to DAILY
    const [projectActive, setIsProjectActive] = useState(project?.isActive);
    const [deactivationConfirmOpen, setDeactivationConfirmOpen] = useState(false);

    const handleMaxCommentsInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setMaxCommentsInput(e.target.value.replace(/\D/g, '')); // allows only digits
    };

    useEffect(() => {
        // In a real application, you would fetch these settings from your backend
        // For example:
        // portalClient.getNotificationSettings().then(res => {
        //   setEmailFrequency(res.frequency || 'DAILY');
        //   setIsProjectDeactivated(res.isDeactivated || false);
        // });

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

    const [showCookieModal, setShowCookieModal] = useState(false);
    const [cookieInput, setCookieInput] = useState('');
    const [cookieError, setCookieError] = useState('');
    const [isSubmittingCookie, setIsSubmittingCookie] = useState(false);


    const handleConnectReddit = async () => {
        let popup: Window | null = null;
        try {
            setIsConnecting(true);
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

            await handleSaveAutomation({ dm: { enabled: true } });

            // Refresh integrations to reflect the newly connected status
            const res = await portalClient.getIntegrations({});
            setIntegrations(res.integrations);

            toast.success("Reddit connected successfully");
        } catch (err: any) {
            if (popup && !popup.closed) {
                popup.close();
            }
            setShowCookieModal(true); // show modal
            // toast.error(getConnectError(err));
        } finally {
            setIsConnecting(false);
        }
    };

    const handleDisconnectReddit = (id: string) => {
        // Optimistically update the state to AUTH_REMOVED
        setIntegrations((prev) =>
            prev.map((i) =>
                i.id === id ? { ...i, status: IntegrationState.AUTH_REVOKED } : i
            )
        );

        // Send API call async
        portalClient.revokeIntegration({ id: id })
            .then(() => {
                // After successful revoke, check current state
                setIntegrations((prev) => {
                    const updated = prev.map((i) =>
                        i.id === id ? { ...i, status: IntegrationState.AUTH_REVOKED } : i
                    );

                    const hasActiveReddit = updated.some(
                        (i) => i.type === IntegrationType.REDDIT_DM_LOGIN && i.status === IntegrationState.ACTIVE
                    );

                    if (!hasActiveReddit) {
                        console.log("No active reddit DM account, disabling automated DMs");
                        handleSaveAutomation({ dm: { enabled: false } });
                    }

                    return updated;
                });
            })
            .catch((err) => {
                toast.error(getConnectError(err));
            });
    };

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
                if (!prev) return prev;
                const updatedOrganizations = prev.organizations.map(org =>
                    org.id === result.id ? result : org
                );
                return { ...prev, organizations: updatedOrganizations };
            });

            toast.success("Automation settings updated successfully!");
        } catch (err) {
            if (err instanceof Error) {
                const message = err.message || "Failed to update automation settings";
                toast.error(message);
            } else {
                console.error("Unexpected error:", err);
            }
        }
    };

    const handleSaveAutomatedComments = (newAutoComment?: boolean) => {
        const maxPerDay = parseInt(maxCommentsInput, 10);
        const enabled = typeof newAutoComment === 'boolean' ? newAutoComment : autoComment;

        if (maxPerDay > 0 && maxPerDay <= maxCommentPerDayLimit) {
            handleSaveAutomation({
                comment: {
                    enabled,
                    relevancyScore,
                    maxPerDay: BigInt(maxPerDay),
                },
            });
        } else {
            toast.error(`Max comments per day must be between 1 and ${maxCommentPerDayLimit}`);
        }
    };

    // New function to handle notification frequency change and auto-save
    const handleEmailFrequencyChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const stringValue = event.target.value; // This will be "DAILY" or "WEEKLY" (string)
        let newFrequencyEnumValue: NotificationFrequency;

        // Explicitly map the string value to the correct Protobuf enum value
        switch (stringValue) {
            case 'DAILY':
                newFrequencyEnumValue = NotificationFrequency.DAILY;
                break;
            case 'WEEKLY':
                newFrequencyEnumValue = NotificationFrequency.WEEKLY;
                break;
            // Add other cases if you have more frequencies in your enum
            default:
                console.warn(`Unknown email frequency string received: ${stringValue}`);
                toast.error("Invalid frequency selected.");
                return; // Stop execution if value is unexpected
        }

        setEmailFrequency(newFrequencyEnumValue);

        try {
            await portalClient.updateAutomationSettings({
                notificationSettings: {
                    relevantPostFrequency: newFrequencyEnumValue
                }
            });
            toast.success("Notification frequency updated!");
        } catch (err) {
            if (err instanceof Error) {
                toast.error(err.message || "Failed to update notification settings");
            } else {
                console.error("Unexpected error:", err);
            }
            // Revert to previous frequency on error, or handle as per UX needs
            setEmailFrequency(emailFrequency); // Revert on error
        }
    };


    const handleProjectStatusUpdate = async (isActive: boolean) => {
        setIsProjectActive(isActive);
        if (!isActive) {
            setDeactivationConfirmOpen(false);
        }
        try {
            await portalClient.updateAutomationSettings({
                projectActive: isActive
            });
            if (isActive) {
                toast.success("Project Activated Successfully");
            } else {
                toast.success("Project Deactivated Successfully");
            }

        } catch (err) {
            setIsProjectActive(!isActive);
            if (err instanceof Error) {
                toast.error(err.message || "Failed to update notification settings");
            } else {
                console.error("Unexpected error:", err);
            }
        }
    };

    if (loading) {
        return <FallbackSpinner />;
    }

    const redditDMIntegrations = integrations.filter(
        (i) => i.type === IntegrationType.REDDIT_DM_LOGIN
    );

    return (
        // Added padding to the main Box
        <Box component="main" sx={{ flexGrow: 1, p: 3, display: "flex", flexDirection: "column" }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                <Tabs value={currentTab} onChange={(_event, newValue) => setCurrentTab(newValue)} aria-label="settings tabs">
                    <Tab label="Automation Settings" />
                    <Tab label="Notification Settings" />
                </Tabs>
            </Box>

            {/* Added padding to the content area below tabs */}
            <Box sx={{ p: { xs: 1, sm: 3 }, flexGrow: 1 }}>
                {currentTab === 0 && (
                    <>
                        {/* DM automation settings */}
                        <Card sx={{ p: 2, mt: 5 }} component={Paper}>
                            <CardContent>

                                <Dialog open={showCookieModal}
                                    onClose={(_, reason) => {
                                        if (reason !== 'backdropClick' && reason !== 'escapeKeyDown') {
                                            setShowCookieModal(false);
                                        }
                                    }}
                                    disableEscapeKeyDown
                                    maxWidth="sm"
                                    fullWidth>
                                    <DialogContent>
                                        <Card variant="outlined" sx={{ backgroundColor: "#f9f9f9", borderRadius: 2, p: 2, mb: 5 }}>
                                            <CardContent>
                                                <Typography variant="subtitle1" gutterBottom fontWeight="bold">
                                                    Connect Reddit Account Manually
                                                </Typography>

                                                <Typography variant="body2" sx={{ mb: 2 }}>
                                                    We couldn't connect to Reddit automatically. Please follow the steps below to connect your account manually:
                                                </Typography>

                                                <List sx={{ listStyleType: 'disc', pl: 3 }} disablePadding>
                                                    <ListItem sx={{ display: 'list-item', py: 0.5 }}>
                                                        <ListItemText
                                                            primary={
                                                                <Typography variant="body2">
                                                                    Go to{" "}
                                                                    <Link href="https://www.reddit.com" target="_blank" rel="noopener" sx={{ textDecoration: "underline" }}>
                                                                        reddit.com
                                                                    </Link>{" "}
                                                                    and log in to your Reddit account.
                                                                </Typography>
                                                            }
                                                        />
                                                    </ListItem>

                                                    <ListItem sx={{ display: 'list-item', py: 0.5 }}>
                                                        <ListItemText
                                                            primary={
                                                                <Typography variant="body2">
                                                                    Install the Chrome extension{" "}
                                                                    <Link
                                                                        href="https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm"
                                                                        target="_blank"
                                                                        rel="noopener"
                                                                        sx={{ textDecoration: "underline" }}
                                                                    >
                                                                        EditThisCookie
                                                                    </Link>
                                                                    .
                                                                </Typography>
                                                            }
                                                        />
                                                    </ListItem>

                                                    <ListItem sx={{ display: 'list-item', py: 0.5 }}>
                                                        <ListItemText
                                                            primary={
                                                                <Typography variant="body2">
                                                                    Open the extension and copy all cookies for reddit.com in JSON format.
                                                                </Typography>
                                                            }
                                                        />
                                                    </ListItem>

                                                    <ListItem sx={{ display: 'list-item', py: 0.5 }}>
                                                        <ListItemText
                                                            primary={
                                                                <Typography variant="body2">
                                                                    Paste the copied cookie JSON into the field below and click <strong>Submit</strong>.
                                                                </Typography>
                                                            }
                                                        />
                                                    </ListItem>
                                                </List>
                                            </CardContent>
                                        </Card>



                                        <TextField
                                            multiline
                                            fullWidth
                                            rows={5}
                                            label="Reddit Cookies JSON"
                                            value={cookieInput}
                                            onChange={(e) => setCookieInput(e.target.value)}
                                            error={!!cookieError}
                                            helperText={cookieError || 'Paste your cookies in JSON format. Validation will take a few mins'}
                                        />
                                    </DialogContent>
                                    <DialogActions>
                                        <Button onClick={() => setShowCookieModal(false)} disabled={isSubmittingCookie}>
                                            Cancel
                                        </Button>
                                        <LoadingButton
                                            loading={isSubmittingCookie}
                                            variant="contained"
                                            onClick={async () => {
                                                try {
                                                    setIsSubmittingCookie(true);
                                                    setCookieError('');

                                                    // ✅ Validate JSON format
                                                    try {
                                                        JSON.parse(cookieInput);
                                                    } catch (e) {
                                                        setCookieError('Please enter valid JSON format.');
                                                        return;
                                                    }

                                                    const response = portalClient.connectReddit(
                                                        {
                                                            cookieJson: cookieInput
                                                        }
                                                    );

                                                    for await (const msg of response) { }

                                                    await handleSaveAutomation({ dm: { enabled: true } });
                                                    const res = await portalClient.getIntegrations({});
                                                    setIntegrations(res.integrations);
                                                    toast.success("Reddit connected successfully");
                                                    setShowCookieModal(false);
                                                    setCookieInput('');
                                                } catch (e: any) {
                                                    setCookieError(getConnectError(e));
                                                } finally {
                                                    setIsSubmittingCookie(false);
                                                }
                                            }}
                                        >
                                            Submit
                                        </LoadingButton>
                                    </DialogActions>
                                </Dialog>

                                <Box display="flex" alignItems="center" gap={1} mb={2}>
                                    <Typography variant="h5" fontWeight="bold">
                                        DM Automation Settings
                                    </Typography>
                                </Box>

                                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                                    {`Redora will auto-send up to ${maxDMPerDay} DMs daily to qualified leads. Your credentials are never stored — we use browser cookies to simulate real user behavior when sending DMs.`}
                                </Typography>

                                <Typography variant="body2" color="text.secondary">
                                    You should log in using your Reddit email (or username) and password. If your account doesn’t have a password set, follow the guide below to set one:
                                </Typography>

                                <Typography sx={{ mt: 2 }} variant="body2" color="primary">
                                    <Link href="https://redoraai.featurebase.app/help/articles/9204295" target="_blank" rel="noopener">
                                        How do I add a password to my account? — Reddit Help
                                    </Link>
                                </Typography>

                                <Box sx={{ mt: 5 }}>
                                    <Paper
                                        variant="outlined"
                                        sx={{
                                            p: 2,
                                            mb: 2,
                                            borderLeft: "4px solid #ff4500",
                                            backgroundColor: "#fff8f6",
                                        }}
                                    >
                                        <Typography variant="body2" sx={{ color: "#4d2c19" }}>
                                            <strong>Note:</strong> You can connect multiple Reddit accounts. We will automatically rotate between them when sending DMs.
                                        </Typography>
                                    </Paper>
                                    {redditDMIntegrations.length > 0 && (
                                        <TableContainer component={Paper} variant="outlined" sx={{ mt: 2 }}>
                                            <Table>
                                                <TableHead>
                                                    <TableRow>
                                                        <TableCell sx={{ fontWeight: "medium" }}>Username</TableCell>
                                                        <TableCell sx={{ fontWeight: "medium" }}>State</TableCell>
                                                        <TableCell />
                                                    </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                    {redditDMIntegrations.map((integration) => {
                                                        const username = integration.details?.value?.userName ?? '—';
                                                        const isActive = integration.status === IntegrationState.ACTIVE;
                                                        const isAuthRemoved = integration.status !== IntegrationState.ACTIVE;
                                                        const reason = integration.details?.value?.reason;

                                                        return (
                                                            <TableRow key={integration.id}>
                                                                <TableCell>
                                                                    {username !== '—' ? (
                                                                        <div className="flex flex-col">
                                                                            <a
                                                                                href={`https://reddit.com/user/${username}`}
                                                                                target="_blank"
                                                                                rel="noopener noreferrer"
                                                                                className="text-blue-600 hover:underline"
                                                                            >
                                                                                {username}
                                                                            </a>
                                                                            {reason && (
                                                                                <span className="text-xs text-yellow-700 mt-1">
                                                                                    {reason}
                                                                                </span>
                                                                            )}
                                                                        </div>
                                                                    ) : (
                                                                        '—'
                                                                    )}
                                                                </TableCell>
                                                                <TableCell>
                                                                    {isActive ? 'Active' : 'Revoked'}
                                                                </TableCell>
                                                                <TableCell align="right">
                                                                    {isActive ? (
                                                                        <Button
                                                                            variant="outlined"
                                                                            color="error"
                                                                            size="small"
                                                                            onClick={() => handleDisconnectReddit(integration.id)}
                                                                        >
                                                                            Disconnect
                                                                        </Button>
                                                                    ) : isAuthRemoved ? (
                                                                        <Button
                                                                            variant="outlined"
                                                                            size="small"
                                                                            onClick={() => handleConnectReddit()}
                                                                        >
                                                                            Reconnect
                                                                        </Button>
                                                                    ) : null}
                                                                </TableCell>
                                                            </TableRow>
                                                        );
                                                    })}
                                                </TableBody>
                                            </Table>
                                        </TableContainer>
                                    )}

                                    <SaveButton
                                        onClick={handleConnectReddit}
                                        variant="contained"
                                        size="large"
                                        disabled={isConnecting}
                                    >
                                        {isConnecting ? 'Connecting...' : 'Connect Reddit DM'}
                                    </SaveButton>
                                </Box>

                            </CardContent>
                        </Card>

                        <Card sx={{ p: 2, mt: 5, mb: 10 }} component={Paper} elevation={3}>
                            <CardContent>
                                {/* Title */}
                                <Box display="flex" alignItems="center" gap={1} mb={3}>
                                    <Typography variant="h5" component="div" fontWeight="bold">
                                        Automated Comments Settings
                                    </Typography>
                                </Box>

                                {/* Description */}
                                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                                    Redora will automatically post up to <strong>{maxCommentPerDay}</strong> comments per day on relevant posts to engage qualified leads. Adjust the settings below to fine-tune this automation.
                                </Typography>

                                {/* Relevancy Score */}
                                <Box mb={4}>
                                    <Typography variant="subtitle1" fontWeight="medium" mb={1}>
                                        Minimum Relevancy Score: {relevancyScore}%
                                    </Typography>

                                    <Box px={2}>
                                        <StyledSlider
                                            value={relevancyScore}
                                            onChange={handleScoreChange}
                                            min={80}
                                            max={100}
                                            step={5}
                                            aria-label="Relevancy Score"
                                        />
                                    </Box>

                                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                                        Only comment on posts with a relevancy score ≥ selected threshold.
                                    </Typography>
                                </Box>

                                {/* Max Comments Per Day Input */}
                                <Box mb={4}>
                                    <Typography variant="subtitle1" fontWeight="medium" mb={1}>
                                        Daily Comment Limit
                                    </Typography>

                                    <Box
                                        display="flex"
                                        alignItems="center"
                                        gap={2}
                                        sx={{
                                            border: '1px solid #d1d5db',
                                            borderRadius: '8px',
                                            padding: '10px 16px',
                                            maxWidth: '260px',
                                            backgroundColor: '#f9fafb',
                                        }}
                                    >
                                        <input
                                            type="number"
                                            min={1}
                                            max={Number(maxCommentPerDayLimit)}
                                            value={maxCommentsInput}
                                            onChange={handleMaxCommentsInputChange}
                                            style={{
                                                flex: 1,
                                                padding: '8px 10px',
                                                border: 'none',
                                                outline: 'none',
                                                fontSize: '1rem',
                                                backgroundColor: 'transparent',
                                                color: '#111827',
                                            }}
                                        />
                                        <Typography variant="body2" color="text.secondary">
                                            {`/ ${maxCommentPerDayLimit}`}
                                        </Typography>
                                    </Box>

                                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1, ml: 0.5 }}>
                                        {/* {`Enter a number between 1 and ${maxCommentPerDayLimit}`}. */}
                                        {`Enter a number between 1 and ${maxCommentPerDayLimit}. To stay safe, we recommend starting with 2–5 comments/day to reduce the risk of Reddit bans.`}
                                    </Typography>
                                </Box>

                                {/* Toggle Switch */}
                                <Box display="flex" alignItems="center" py={2} mb={4}>
                                    <CustomSwitch
                                        checked={autoComment}
                                        onChange={(e) => {
                                            const newValue = e.target.checked;
                                            setAutoComment(newValue);
                                            handleSaveAutomatedComments(newValue); // auto-save when toggled
                                        }}
                                    />
                                    <Typography variant="body1" fontWeight="medium" ml={2.5} display="flex">
                                        Automated Comments
                                        <Typography
                                            variant="body1"
                                            fontWeight="medium"
                                            ml={1}
                                            sx={{ color: autoComment ? 'green' : 'red' }}
                                        >
                                            {autoComment ? 'On' : 'Off'}
                                        </Typography>
                                    </Typography>
                                </Box>

                                {/* Save Button */}
                                <SaveButton
                                    onClick={() => handleSaveAutomatedComments()}
                                    variant="contained"
                                    size="large"
                                >
                                    Save Automation Settings
                                </SaveButton>
                            </CardContent>
                        </Card>
                    </>
                )}

                {currentTab === 1 && (
                    <>
                        {/* Notification Settings */}
                        <Card sx={{ p: 4, mt: 5 }} component={Paper}>
                            <CardContent>
                                <Box display="flex" alignItems="center" gap={1} mb={2}>
                                    <Typography variant="h4" fontWeight="bold">
                                        Notification Settings
                                    </Typography>
                                </Box>

                                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                                    Manage how often you receive email alerts for relevant posts.
                                </Typography>

                                <Box mb={4}>
                                    <Typography variant="subtitle1" fontWeight="medium" mb={1}>
                                        Relevant Posts Email Alerts Frequency:
                                    </Typography>
                                    <RadioGroup
                                        row
                                        aria-label="email-frequency"
                                        name="email-frequency-group"
                                        value={NotificationFrequency[emailFrequency]}
                                        onChange={handleEmailFrequencyChange}
                                    >
                                        <FormControlLabel value="DAILY" control={<Radio />} label="Daily" />
                                        <FormControlLabel value="WEEKLY" control={<Radio />} label="Weekly" />
                                    </RadioGroup>
                                </Box>

                                {/* Removed the "Save Notification Settings" button */}

                            </CardContent>
                        </Card>

                        {/* Deactivate Project */}
                        <Card sx={{ p: 4, mt: 5, mb: 10 }} component={Paper} elevation={3}>
                            <CardContent>
                                <Box display="flex" alignItems="center" gap={1} mb={3}>
                                    <Typography variant="h4" component="div" fontWeight="bold">
                                        Project Status
                                    </Typography>
                                </Box>

                                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                                    {!projectActive
                                        ? "This project is currently deactivated. Automated activities and email notifications are paused. You can reactivate it at any time."
                                        : "Deactivating your project will stop all automated activities and email notifications. This action can be reversed by reactivating the project."
                                    }
                                </Typography>

                                {!projectActive ? (
                                    <SaveButton
                                        variant="contained"
                                        disabled={hasPlanExpired}
                                        size="large"
                                        onClick={() => handleProjectStatusUpdate(true)} // Calls new reactivation function
                                    >
                                        Reactivate Project
                                    </SaveButton>
                                ) : (
                                    <Button
                                        variant="contained"
                                        disabled={hasPlanExpired}
                                        color="error"
                                        size="large"
                                        onClick={() => setDeactivationConfirmOpen(true)}
                                    >
                                        Deactivate Project
                                    </Button>
                                )}
                            </CardContent>
                        </Card>
                    </>
                )}
            </Box>

            {/* Deactivation Confirmation Dialog */}
            <Dialog
                open={deactivationConfirmOpen}
                onClose={() => setDeactivationConfirmOpen(false)}
                aria-labelledby="deactivation-dialog-title"
                aria-describedby="deactivation-dialog-description"
            >
                <DialogTitle id="deactivation-dialog-title">{"Confirm Project Deactivation"}</DialogTitle>
                <DialogContent>
                    <Typography id="deactivation-dialog-description">
                        Are you sure you want to deactivate this project? All automated activities will stop. You can reactivate it later.
                    </Typography>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setDeactivationConfirmOpen(false)}>Cancel</Button>
                    <Button onClick={() => handleProjectStatusUpdate(false)} color="error" autoFocus>
                        Deactivate
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
}