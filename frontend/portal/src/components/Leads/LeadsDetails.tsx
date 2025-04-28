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
import { ThumbDown, ThumbUp, Close, Star, Send } from "@mui/icons-material"
import { LightbulbIcon } from "lucide-react";
import { formateDate, getSubredditName } from "./Tabs/NewTab";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import Link from "next/link";
import React from 'react';
import ReactMarkdown from 'react-markdown';
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { setSelectedLeadData } from "../../../store/Lead/leadSlice";

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

const LeadsPostDetails = () => {

  const dispatch = useAppDispatch();
  const { subredditList } = useAppSelector((state: RootState) => state.source);
  const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

  const { portalClient } = useClientsContext();
  const [isLoading, setIsLoading] = useState<boolean>(false);

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
    dispatch(setSelectedLeadData(null));
  };

  const handleLeadNotRelevent = async () => {
    setIsLoading(true);
    if (!selectedleadData) return;
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
    if (!selectedleadData) return;
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

  const decodeHtml = (html: string) => {
    const textarea = document.createElement('textarea');
    textarea.innerHTML = html;
    return textarea.value;
  };

  // console.log("###_leads ", selectedleadData);

  if (!selectedleadData) return null;

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
                  <Link href={selectedleadData.metadata?.authorUrl as string} target="_blank">{selectedleadData.author}</Link> • {getSubredditName(subredditList, selectedleadData.sourceId)} • {selectedleadData.postCreatedAt ? formateDate(selectedleadData.postCreatedAt) : "N/A"}
                </Typography>
                <Link href={selectedleadData.metadata?.postUrl as string} target="_blank">
                  <Typography variant="h5" component="h1" sx={{ fontWeight: "bold", mb: 2 }}>
                    {selectedleadData.title}
                  </Typography>
                </Link>
                <div dangerouslySetInnerHTML={{ __html: decodeHtml(selectedleadData.metadata?.descriptionHtml as string) }} />
              </Box>
            </Box>

            {/* Suggested comment card */}
            {selectedleadData.metadata?.suggestedComment && (
              <Card sx={{ mb: 2, borderRadius: 2, mx: 2 }}>
                <CardContent>
                  <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                    <Star sx={{ color: "#e25a9e", mr: 1 }} />
                    <Typography color="#e25a9e" fontWeight="medium">
                      Suggested comment
                    </Typography>
                  </Box>
                  <Typography variant="body1" sx={{ mb: 2 }}>
                    {selectedleadData.metadata?.suggestedComment}
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
                      onClick={() => copyTextAndOpenLink(selectedleadData.metadata?.suggestedComment as string, selectedleadData.metadata?.postUrl as string)}
                    >
                      Copy & open post
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            )}

            {/* Suggested DM card */}
            {selectedleadData.metadata?.suggestedDm && (
              <Card sx={{ borderRadius: 2, mx: 2 }}>
                <CardContent>
                  <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                    <Star sx={{ color: "#e25a9e", mr: 1 }} />
                    <Typography color="#e25a9e" fontWeight="medium">
                      Suggested DM
                    </Typography>
                  </Box>
                  <Typography variant="body1" paragraph>
                    {selectedleadData.metadata?.suggestedDm}
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
                      onClick={() => copyTextAndOpenLink(selectedleadData.metadata?.suggestedDm as string, selectedleadData.metadata?.dmUrl as string)}
                    >
                      Copy & open DMs
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            )}
          </Box>

        </Paper>
      </Box>
    </ThemeProvider>
  )
}

export default LeadsPostDetails;
