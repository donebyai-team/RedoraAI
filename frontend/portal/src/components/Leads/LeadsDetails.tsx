"use client";

import { useState } from "react"
import {
  Box,
  Typography,
  Paper,
  IconButton,
  Button,
  Chip,
  ThemeProvider,
  createTheme,
  Card,
  CardContent,
  Stack,
  Tooltip,
} from "@mui/material"
import { ThumbDown, ThumbUp, Close, Lightbulb, Star, Refresh, Edit, Send } from "@mui/icons-material"

// Create a custom theme with Reddit-like colors
const theme = createTheme({
  palette: {
    primary: {
      main: "#ff4500", // Reddit orange
    },
    secondary: {
      main: "#0079d3", // Reddit blue
    },
    background: {
      default: "#dae0e6",
      paper: "#ffffff",
    },
  },
  typography: {
    fontFamily: '"Noto Sans", "Helvetica", "Arial", sans-serif',
  },
})

export default function LeadsPostDetails() {
  const [relevancy] = useState(70)

  return (
    <ThemeProvider theme={theme}>
      <Box sx={{ width: "100%", }}>
        <Paper elevation={0} sx={{ borderRadius: 2, overflow: "hidden", mb: 2 }}>
          {/* Header with relevancy and actions */}
          <Box
            sx={{
              display: "flex",
              justifyContent: "space-between",
              p: 1.5,
              borderBottom: "1px solid #f0f0f0",
            }}
          >
            <Box sx={{ display: "flex", alignItems: "center" }}>
              <Chip
                icon={<Lightbulb sx={{ color: "#b07d1a !important" }} />}
                label={`${relevancy}% relevancy`}
                sx={{
                  bgcolor: "rgba(255, 215, 0, 0.2)",
                  color: "#b07d1a",
                  fontWeight: "bold",
                  "& .MuiChip-icon": { color: "#b07d1a" },
                }}
              />
            </Box>
            <Box sx={{ display: "flex", gap: 1 }}>
              <Button
                variant="contained"
                startIcon={<ThumbDown />}
                sx={{
                  bgcolor: "#f0f0f0",
                  color: "#666",
                  "&:hover": { bgcolor: "#e0e0e0" },
                  textTransform: "none",
                  boxShadow: "none",
                }}
              >
                Not relevant
              </Button>
              <Button
                variant="contained"
                startIcon={<ThumbUp />}
                sx={{
                  bgcolor: "#f0f0f0",
                  color: "#666",
                  "&:hover": { bgcolor: "#e0e0e0" },
                  textTransform: "none",
                  boxShadow: "none",
                }}
              >
                Complete
              </Button>
              <IconButton size="small">
                <Close />
              </IconButton>
            </Box>
          </Box>

          {/* Post content */}
          <Box sx={{ p: 2 }}>
            <Box sx={{ mb: 1 }}>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                /u/Action_Hank1 • r/sales • 30 minutes ago
              </Typography>
              <Typography variant="h5" component="h1" sx={{ fontWeight: "bold", mb: 2 }}>
                Is your manager getting their playbook from LinkedIn Influencers?
              </Typography>
              <Typography variant="body1" paragraph>
                Title - you know those posts by sales influencers that break down some too good to be true scenario with
                a CTA that involves you commenting a one word answer like "playbook" or "outbound" on their post to get
                access to their secret (aka get put in their sales funnel)?
              </Typography>
              <Typography variant="body1" paragraph>
                I've seen a few former colleagues who are sales managers getting sucked into these posts.
              </Typography>
              <Typography variant="body1" paragraph>
                On one hand, I think that if I saw my manager looking for answers on LinkedIn, I'd wonder about my
                management team's competence.
              </Typography>
              <Typography variant="body1" paragraph>
                On the other hand, maybe you need to go external for more knowledge. Sales is hard and a lot of places
                are struggling right now. Especially if you're a nice-to-have product/service.
              </Typography>
            </Box>
          </Box>
        </Paper>

        {/* Suggested comment card */}
        <Card sx={{ mb: 2, borderRadius: 2, mx: 2 }}>
          <CardContent>
            <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
              <Star sx={{ color: "#e25a9e", mr: 1 }} />
              <Typography color="#e25a9e" fontWeight="medium">
                Suggested comment
              </Typography>
            </Box>
            <Typography variant="body1" sx={{ mb: 2 }}>
              Sales gurus are sus. Managers should learn from real work.
            </Typography>
            <Stack direction="row" justifyContent="space-between">
              <Box>
                <Tooltip title="Rewrite">
                  <IconButton>
                    <Refresh fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="Edit prompt">
                  <IconButton>
                    <Edit fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Box>
              <Button
                variant="contained"
                color="primary"
                startIcon={<Send />}
                sx={{
                  bgcolor: "#000",
                  color: "#fff",
                  "&:hover": { bgcolor: "#333" },
                  borderRadius: "20px",
                  textTransform: "none",
                }}
              >
                Copy & open post
              </Button>
            </Stack>
          </CardContent>
        </Card>

        {/* Suggested DM card */}
        <Card sx={{ borderRadius: 2, mx: 2 }}>
          <CardContent>
            <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
              <Star sx={{ color: "#e25a9e", mr: 1 }} />
              <Typography color="#e25a9e" fontWeight="medium">
                Suggested DM
              </Typography>
            </Box>
            <Typography variant="body1" paragraph>
              Hi Hank, I'm John from Acme. I saw your post about managers getting sales advice from LinkedIn. I think
              you could really use Acme to automate your team's sales tasks. Acme is an AI agent that can automate any
              task. You can try it for free at acme.com
            </Typography>
            <Stack direction="row" justifyContent="space-between">
              <Box>
                <Tooltip title="Rewrite">
                  <IconButton>
                    <Refresh fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="Edit prompt">
                  <IconButton>
                    <Edit fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Box>
              <Button
                variant="contained"
                color="primary"
                startIcon={<Send />}
                sx={{
                  bgcolor: "#000",
                  color: "#fff",
                  "&:hover": { bgcolor: "#333" },
                  borderRadius: "20px",
                  textTransform: "none",
                }}
              >
                Copy & open DMs
              </Button>
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </ThemeProvider>
  )
}
