"use client";

import { useState, useMemo, useCallback, memo } from "react";
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
} from "@mui/material";
import { ThumbDown, ThumbUp, Close, Star, Send } from "@mui/icons-material";
import { LightbulbIcon } from "lucide-react";
import Link from "next/link";
import toast from "react-hot-toast";

import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import {
  LeadTabStatus,
  setCompletedList,
  setDiscardedTabList,
  setNewTabList,
  setSelectedLeadData,
} from "../../../store/Lead/leadSlice";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { HtmlTitleRenderer, HtmlBodyRenderer, MarkdownRenderer } from "../Html/HtmlRenderer";
import { formateDate, getSubredditName } from "./Tabs/NewTab";

// Memoized renderers
const MemoizedHtmlTitleRenderer = memo(HtmlTitleRenderer);
const MemoizedHtmlBodyRenderer = memo(HtmlBodyRenderer);

// Memoized theme
const redditTheme = createTheme({
  palette: {
    primary: { main: "#ff4500" },
    secondary: { main: "#0079d3" },
    background: { default: "#dae0e6", paper: "#ffffff" },
  },
  typography: {
    fontFamily: '"Noto Sans", "Helvetica", "Arial", sans-serif',
  },
});

const LeadsPostDetails = () => {
  const dispatch = useAppDispatch();
  const { portalClient } = useClientsContext();

  const selectedleadData = useAppSelector((state: RootState) => state.lead.selectedleadData);
  const activeTab = useAppSelector((state: RootState) => state.lead.activeTab);
  const newTabList = useAppSelector((state: RootState) => state.lead.newTabList);
  const completedTabList = useAppSelector((state: RootState) => state.lead.completedTabList);
  const discardedTabList = useAppSelector((state: RootState) => state.lead.discardedTabList);
  const subredditList = useAppSelector((state: RootState) => state.source.subredditList);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const subredditName = useMemo(() => {
    return getSubredditName(subredditList, selectedleadData?.sourceId ?? "");
  }, [subredditList, selectedleadData?.sourceId]);

  const formattedDate = useMemo(() => {
    return selectedleadData?.postCreatedAt ? formateDate(selectedleadData.postCreatedAt) : "N/A";
  }, [selectedleadData?.postCreatedAt]);

  const handleCloseLeadDetail = useCallback(() => {
    dispatch(setSelectedLeadData(null));
  }, [dispatch]);

  const handleSelectNext = useCallback((status: LeadStatus) => {
    if (activeTab !== LeadTabStatus.NEW || !selectedleadData) return;

    const currentIndex = newTabList.findIndex(item => item.id === selectedleadData.id);
    const nextItem = newTabList[currentIndex + 1];
    const newTabListArray = newTabList.filter(item => item.id !== selectedleadData.id);

    if (status === LeadStatus.COMPLETED) {
      dispatch(setCompletedList([...completedTabList, selectedleadData]));
    } else if (status === LeadStatus.NOT_RELEVANT) {
      dispatch(setDiscardedTabList([...discardedTabList, selectedleadData]));
    }

    if (nextItem) {
      dispatch(setSelectedLeadData(nextItem));
      dispatch(setNewTabList(newTabListArray));
    } else {
      handleCloseLeadDetail();
    }
  }, [activeTab, completedTabList, discardedTabList, dispatch, newTabList, selectedleadData, handleCloseLeadDetail]);

  const copyTextAndOpenLink = useCallback((textToCopy: string, linkToOpen: string) => {
    if (!navigator.clipboard) {
      // Fallback for older browsers that do not support `navigator.clipboard`
      const textArea = document.createElement("textarea");
      textArea.value = textToCopy;
      textArea.style.position = "fixed";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        const successful = document.execCommand("copy");
        if (!successful) throw new Error("Fallback: Copy command was unsuccessful");
        window.open(linkToOpen, '_blank');
      } catch (err: any) {
        const message = err?.message || "Fallback: Copy failed";
        toast.error(message);
      } finally {
        document.body.removeChild(textArea);
      }
    } else {
      navigator.clipboard.writeText(textToCopy)
        .then(() => window.open(linkToOpen, '_blank'))
        .catch((err: any) => {
          const message = err?.message || "Clipboard copy failed";
          toast.error(message);
        });
    }
  }, []);

  const handleLeadStatusUpdate = useCallback(async (status: LeadStatus) => {
    if (!selectedleadData) return;
    setIsLoading(true);
    try {
      await portalClient.updateLeadStatus({ status, leadId: selectedleadData.id });
      handleSelectNext(status);
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, [portalClient, selectedleadData, handleSelectNext]);

  if (!selectedleadData) return null;

  return (
    <ThemeProvider theme={redditTheme}>
      <Box sx={{ width: "100%" }}>
        <Paper elevation={0} sx={{ borderRadius: 2, overflow: "hidden" }}>
          {/* Header */}
          <Box
            sx={{
              display: "flex",
              justifyContent: "space-between",
              px: 1.5,
              py: 1.4,
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
                }}
              />


              {selectedleadData.metadata?.relevancyLlmModel && (
                <Chip
                  label={`${selectedleadData.metadata.relevancyLlmModel}${selectedleadData.metadata.llmModelResponseOverriddenBy
                    ? `\n${selectedleadData.metadata.llmModelResponseOverriddenBy}`
                    : ""
                    }`}
                  sx={{
                    whiteSpace: "pre-line", // allows \n to render as line break
                    bgcolor: "rgba(0, 123, 255, 0.1)",
                    color: "#0056b3",
                    fontWeight: "bold",
                  }}
                />
              )}



              <Tooltip
                title={
                  <Box>
                    <MarkdownRenderer data={selectedleadData.metadata?.chainOfThought || ""} />
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
                  },
                }}
              >
                <IconButton sx={{ borderRadius: 2.5 }}>
                  <LightbulbIcon size={22} />
                </IconButton>
              </Tooltip>
            </Box>
            <Box sx={{ display: "flex", gap: 1 }}>
              {activeTab === LeadTabStatus.NEW && (
                <>
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
                    onClick={() => handleLeadStatusUpdate(LeadStatus.NOT_RELEVANT)}
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
                    onClick={() => handleLeadStatusUpdate(LeadStatus.COMPLETED)}
                    disabled={isLoading}
                  >
                    Mark Responded
                  </Button>
                </>
              )}
              <IconButton size="small" onClick={handleCloseLeadDetail}>
                <Close />
              </IconButton>
            </Box>
          </Box>

          {/* Body */}
          <Box sx={{ width: "100%", pb: 2 }}>
            <Box sx={{ px: 2, pt: 2, height: "42dvh", maxHeight: "100%", overflowY: "scroll" }}>
              {/* Metadata line */}
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                <Link href={selectedleadData.metadata?.authorUrl || "#"} target="_blank">
                  {selectedleadData.author}
                </Link>{" "}
                • {subredditName} • {formattedDate}
              </Typography>

              {/* Title */}
              <Typography
                variant="h5"
                component={Link}
                href={selectedleadData.metadata?.postUrl || "#"}
                target="_blank"
                sx={{ fontWeight: "bold", mb: 2, display: "block", textDecoration: "none" }}
              >
                <MemoizedHtmlTitleRenderer htmlString={selectedleadData.title || ""} />
              </Typography>

              {/* Description with improved styling */}
              <Box sx={{ typography: "body1", lineHeight: 1.7, "& p": { mb: 2 }, "& li": { mb: 1, ml: 3 } }}>
                <MemoizedHtmlBodyRenderer htmlString={selectedleadData.metadata?.descriptionHtml || ""} />
              </Box>
            </Box>
            {(selectedleadData.metadata?.suggestedComment || selectedleadData.metadata?.suggestedDm) && (
              <Stack direction={{ xs: "column", md: "row" }} spacing={2} sx={{ mx: 2, my: 2 }}>
                {/* Suggested Comment */}
                {selectedleadData.metadata?.suggestedComment && (
                  <Card sx={{ flex: 1, borderRadius: 2, bgcolor: "#f3f4f6", display: "flex", flexDirection: "column" }}>
                    <CardContent
                      sx={{
                        flexGrow: 1,
                        display: "flex",
                        flexDirection: "column",
                        justifyContent: "space-between",
                        p: 1.5,
                        "&:last-child": { pb: 1.5 },
                      }}
                    >
                      <Box>
                        <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                          <Star sx={{ color: "#e25a9e", mr: 1 }} />
                          <Typography color="#e25a9e" fontWeight="medium">Suggested comment</Typography>
                        </Box>
                        <Typography variant="body1" sx={{ mb: 2 }}>
                          <MarkdownRenderer data={selectedleadData.metadata?.suggestedComment || ""} />
                        </Typography>
                      </Box>

                      <Stack direction="row" justifyContent="flex-start">
                        <Button
                          variant="contained"
                          startIcon={<Send />}
                          sx={{
                            bgcolor: selectedleadData.metadata?.automatedCommentUrl ? "green" : "#000",
                            color: "#fff",
                            "&:hover": {
                              bgcolor: selectedleadData.metadata?.automatedCommentUrl ? "darkgreen" : "#333",
                            },
                            borderRadius: "20px",
                            textTransform: "none",
                          }}
                          onClick={() =>
                            copyTextAndOpenLink(
                              selectedleadData.metadata?.suggestedComment ?? "",
                              selectedleadData.metadata?.automatedCommentUrl || selectedleadData.metadata?.postUrl || "#"
                            )
                          }
                        >
                          {selectedleadData.metadata?.automatedCommentUrl ? "Commented by AI" : "Copy & open post"}
                        </Button>
                      </Stack>
                    </CardContent>
                  </Card>
                )}

                {/* Suggested DM */}
                {selectedleadData.metadata?.suggestedDm && (
                  <Card sx={{ flex: 1, borderRadius: 2, bgcolor: "#f3f4f6" }}>
                    <CardContent sx={{ p: 1.5, "&:last-child": { pb: 1.5 } }}>
                      <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                        <Star sx={{ color: "#e25a9e", mr: 1 }} />
                        <Typography color="#e25a9e" fontWeight="medium">Suggested DM</Typography>
                      </Box>
                      <Typography variant="body1" sx={{ mb: 2 }}>
                        <MarkdownRenderer data={selectedleadData.metadata?.suggestedDm || ""} />
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
                          onClick={() =>
                            copyTextAndOpenLink(
                              selectedleadData.metadata?.suggestedDm ?? "",
                              selectedleadData.metadata?.dmUrl ?? "#"
                            )
                          }
                        >
                          Copy & open DMs
                        </Button>
                      </Stack>
                    </CardContent>
                  </Card>
                )}
              </Stack>
            )}
          </Box>
        </Paper>
      </Box>
    </ThemeProvider>
  );
};

export default memo(LeadsPostDetails);
