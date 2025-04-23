"use client"

import React, { useState } from "react";
import {
  Box,
  Typography,
  Tabs,
  Tab,
  List,
  ListItem,
  Divider,
  Stack,
} from "@mui/material";

interface Post {
  progress: string;
  progressColor: string;
  subreddit: string;
  timeAgo: string;
  title: string;
}

const mockPosts: Post[] = [
  {
    progress: "100%",
    progressColor: "green",
    subreddit: "r/sales",
    timeAgo: "about 4 hours ago",
    title: "Sales to developing markets",
  },
  {
    progress: "100%",
    progressColor: "green",
    subreddit: "r/sales",
    timeAgo: "about 4 hours ago",
    title: "28 years as an Individual contributor, looking to move to become a sales manager",
  },
  {
    progress: "100%",
    progressColor: "green",
    subreddit: "r/marketing",
    timeAgo: "about 7 hours ago",
    title: "When you gain a new client, do you draw up the contract/agreement or them?",
  },
  {
    progress: "100%",
    progressColor: "green",
    subreddit: "r/marketing",
    timeAgo: "about 8 hours ago",
    title: "Where's the best place to get rack cards, business cards, etc printed?",
  },
  {
    progress: "80%",
    progressColor: "#9ACD32",
    subreddit: "r/marketing",
    timeAgo: "about 11 hours ago",
    title: "Need career path help",
  },
  {
    progress: "100%",
    progressColor: "green",
    subreddit: "r/sales",
    timeAgo: "about 11 hours ago",
    title: "On a 9 month plan to become a sales manager with a new team from being an AE after 10...",
  },
];

const InboxComponent = () => {
  const [tabValue, setTabValue] = useState(0);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    console.log("###_debug_event ", event);
    setTabValue(newValue);
  };

  return (
    <Box
      sx={{
        width: "100%",
        px: 3,
        py: 2,
        maxWidth: "25vw",
        borderRight: "1px solid #e0e0e0",
      }}
    >
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          my: 2,
        }}
      >
        <Typography variant="h4" component="h3" sx={{ fontWeight: "bold" }}>
          Inbox
        </Typography>
      </Box>

      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        sx={{
          mb: 2,
          "& .MuiTab-root": {
            textTransform: "none",
            fontWeight: "medium",
            fontSize: "0.95rem",
            minWidth: "auto",
            px: 2,
          },
          "& .Mui-selected": {
            color: "black",
            fontWeight: "bold",
          },
          "& .MuiTabs-indicator": {
            display: "none",
          },
        }}
      >
        <Tab label="NEW" sx={{ color: "text.secondary" }} />
        <Tab label="COMPLETED" sx={{ color: "text.secondary" }} />
        <Tab label="DISCARDED" sx={{ color: "text.secondary" }} />
      </Tabs>

      {mockPosts.length === 0 ? (
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            height: "60vh",
            textAlign: "center",
            px: 2,
          }}
        >
          <Typography variant="body1" color="text.secondary">
            Sit back and relax, we are finding relevant leads for you. We will
            notify you once it’s ready.
          </Typography>
        </Box>
      ) : (
        <List sx={{ p: 0 }}>
          {mockPosts.map((post, index) => (
            <React.Fragment key={index}>
              <ListItem sx={{ py: 2, px: 3 }}>
                <Stack direction="column" spacing={1} width="100%">
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Box
                      sx={{
                        display: "flex",
                        alignItems: "center",
                        color: post.progressColor,
                        fontSize: "0.875rem",
                      }}
                    >
                      <Box
                        component="span"
                        sx={{
                          display: "inline-block",
                          width: 10,
                          height: 10,
                          borderRadius: "50%",
                          bgcolor: post.progressColor,
                          mr: 0.5,
                        }}
                      />
                      {post.progress}
                    </Box>
                    <Typography
                      component="span"
                      sx={{ fontSize: "0.875rem", mx: 1 }}
                    >
                      •
                    </Typography>
                    <Typography
                      component="span"
                      sx={{
                        fontSize: "0.875rem",
                        color: "text.secondary",
                      }}
                    >
                      {post.subreddit}
                    </Typography>
                    <Typography
                      component="span"
                      sx={{ fontSize: "0.875rem", mx: 1 }}
                    >
                      •
                    </Typography>
                    <Typography
                      component="span"
                      sx={{
                        fontSize: "0.875rem",
                        color: "text.secondary",
                      }}
                    >
                      {post.timeAgo}
                    </Typography>
                  </Stack>
                  <Typography variant="body1" sx={{ fontWeight: "medium" }}>
                    {post.title}
                  </Typography>
                </Stack>
              </ListItem>
              {index !== mockPosts.length - 1 && <Divider />}
            </React.Fragment>
          ))}
        </List>
      )}
    </Box>
  );
};

export default InboxComponent;
