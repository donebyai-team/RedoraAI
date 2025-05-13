"use client";

import { useState } from "react"
import {
    Box,
    Button,
    Typography,
    Paper,
    Grid,
    Alert,
    CircularProgress,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
} from "@mui/material"
import {
    Reddit as RedditIcon,
    Security as SecurityIcon,
    Notifications as NotificationsIcon,
    Analytics as AnalyticsIcon,
} from "@mui/icons-material"

export default function ConnectRedditStep() {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")

    const handleConnectReddit = async () => {
        setLoading(true)
        setError("")

        try {
            // Simulate API call to connect Reddit
            await new Promise((resolve) => setTimeout(resolve, 2000))

            // In a real app, this would redirect to Reddit OAuth
            // and then handle the callback
            // onChange(true)
        } catch (err) {
            setError("Failed to connect to Reddit. Please try again.")
        } finally {
            setLoading(false)
        }
    }

    return (
        <Box>
            {error && (
                <Alert severity="error" sx={{ mb: 3 }}>
                    {error}
                </Alert>
            )}

            <Grid container spacing={4}>
                <Grid item xs={12} md={12}>
                    <Paper
                        variant="elevation"
                        sx={{
                            padding: 5,
                            height: "100%",
                            display: "flex",
                            flexDirection: "column",
                        }}
                        elevation={0}
                    >
                        <Typography variant="h6" gutterBottom>
                            Connect Your Reddit Account
                        </Typography>

                        <Typography variant="body2" color="text.secondary" paragraph>
                            Connecting your Reddit account allows us to monitor discussions about your product and keywords across
                            subreddits.
                        </Typography>

                        <List sx={{ mb: 3 }}>
                            <ListItem>
                                <ListItemIcon>
                                    <SecurityIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Secure OAuth Connection"
                                    secondary="We use Reddit's official OAuth for secure access"
                                />
                            </ListItem>
                            <ListItem>
                                <ListItemIcon>
                                    <NotificationsIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Real-time Notifications"
                                    secondary="Get notified when your keywords are mentioned"
                                />
                            </ListItem>
                            <ListItem>
                                <ListItemIcon>
                                    <AnalyticsIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Detailed Analytics"
                                    secondary="Track engagement and sentiment across platforms"
                                />
                            </ListItem>
                        </List>

                        <Box sx={{ mt: "auto" }}>

                            <Button
                                variant="contained"
                                onClick={handleConnectReddit}
                                disabled={loading}
                                startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <RedditIcon />}
                                fullWidth
                                sx={{
                                    bgcolor: "#FF4500", // Reddit orange
                                    "&:hover": { bgcolor: "#E03D00" },
                                }}
                            >
                                Connect with Reddit
                            </Button>
                        </Box>
                    </Paper>
                </Grid>
            </Grid>
        </Box>
    )
}