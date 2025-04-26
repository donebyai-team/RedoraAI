"use client";

import { useEffect, useState } from "react"
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
import { ThumbDown, ThumbUp, Close, Star, Send } from "@mui/icons-material"
import { ChildComponentProps } from "./Inbox";
import { LightbulbIcon } from "lucide-react";
import { formateDate, getSubredditName } from "./Tabs/NewTab";
import { LeadStatus, SubReddit } from "@doota/pb/doota/reddit/v1/reddit_pb";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import Link from "next/link";
import React from 'react';
import ReactMarkdown from 'react-markdown';

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

const LeadsPostDetails: React.FC<ChildComponentProps> = ({ selectedleadData, setSelectedLeadData }) => {

  if (!selectedleadData) return null;

  const { portalClient } = useClientsContext();
  const [subredditList, setSubredditList] = useState<SubReddit[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {

    const getAllSubReddits = async () => {

      try {
        const result = await portalClient.getSubReddits({});
        setSubredditList(result?.subreddits ?? []);
      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong"
        console.log(message);
      }
    }
    getAllSubReddits();

  }, []);

  const copyTextAndOpenLink = (textToCopy: string, linkToOpen: string) => {
    navigator.clipboard.writeText(textToCopy)
      .then(() => {
        console.log('Text copied successfully!');
        window.open(linkToOpen, '_blank');
      })
      .catch((err) => {
        console.error('Failed to copy text: ', err);
      });
  };

  const handleCloseLeadDetail = () => {
    setSelectedLeadData(null);
  };

  const handleLeadNotRelevent = async () => {
    setIsLoading(true);
    try {
      const result = await portalClient.updateLeadStatus({ status: LeadStatus.NOT_RELEVANT, leadId: selectedleadData.id });
      console.log("###_result ", result);
      handleCloseLeadDetail();
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong"
      console.log("###_error", message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLeadComplete = async () => {
    setIsLoading(true);
    try {
      const result = await portalClient.updateLeadStatus({ status: LeadStatus.COMPLETED, leadId: selectedleadData.id });
      console.log("###_result ", result);
      handleCloseLeadDetail();
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong"
      console.log("###_error", message);
    } finally {
      setIsLoading(false);
    }
  };

  console.log("###_leads ", selectedleadData);

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
            <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
              <Chip
                label={`${selectedleadData.relevancyScore}% relevancy`}
                sx={{
                  bgcolor: "rgba(255, 215, 0, 0.2)",
                  color: "#b07d1a",
                  fontWeight: "bold",
                  "& .MuiChip-icon": { color: "#b07d1a" },
                }}
              />
              <Tooltip
                title={
                  <Box>
                    <ReactMarkdown>{selectedleadData.metadata?.chainOfThought}</ReactMarkdown>
                  </Box>
                }
                placement="bottom-start"
                slotProps={{
                  tooltip: {
                    sx: {
                      backgroundColor: '#fff',
                      color: '#666',
                      boxShadow: 3,
                      borderRadius: 1,
                      p: 1.5,
                      maxWidth: "30vw",
                    },
                  }
                }}
              >
                <IconButton sx={{ borderRadius: 2.5 }}>
                  <LightbulbIcon size={22} />
                </IconButton>
              </Tooltip>
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
                onClick={handleLeadNotRelevent}
                disabled={isLoading}
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
                onClick={handleLeadComplete}
                disabled={isLoading}
              >
                Complete
              </Button>
              <IconButton size="small" onClick={handleCloseLeadDetail}>
                <Close />
              </IconButton>
            </Box>
          </Box>

          <Box sx={{ width: "100%", py: 2, height: "91dvh", overflowY: "scroll" }}>
            {/* Post content */}
            <Box sx={{ p: 2 }}>
              <Box sx={{ mb: 1 }}>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                  <Link href={selectedleadData.metadata?.authorUrl as string} target="_blank">{`/u/Action_Hank1`}</Link> • {getSubredditName(subredditList, selectedleadData.subredditId)} • {selectedleadData.postCreatedAt ? formateDate(selectedleadData.postCreatedAt) : "N/A"}
                </Typography>
                <Typography variant="h5" component="h1" sx={{ fontWeight: "bold", mb: 2 }}>
                  {selectedleadData.title}
                </Typography>
                <ReactMarkdown>{selectedleadData.description}</ReactMarkdown>
              </Box>
            </Box>

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
                  {selectedleadData.metadata?.chainOfThoughtSuggestedComment}
                </Typography>
                <Stack direction="row" justifyContent="end">
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
                    onClick={() => copyTextAndOpenLink(selectedleadData.metadata?.chainOfThoughtSuggestedComment as string, selectedleadData.metadata?.postUrl as string)}
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
                  {selectedleadData.metadata?.chainOfThoughtSuggestedDm}
                </Typography>
                <Stack direction="row" justifyContent="end">
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
                    onClick={() => copyTextAndOpenLink(selectedleadData.metadata?.chainOfThoughtSuggestedDm as string, selectedleadData.metadata?.dmUrl as string)}
                  >
                    Copy & open DMs
                  </Button>
                </Stack>
              </CardContent>
            </Card>
          </Box>

        </Paper>
      </Box>
    </ThemeProvider>
  )
}

export default LeadsPostDetails;
