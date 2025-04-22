"use client"

import type React from "react";
import { useState } from "react";
import { Box, Typography, Button, Tabs, Tab, List, ListItem, Paper, Divider, Stack } from "@mui/material";
import { Edit } from "lucide-react";

const InboxComponent = () => {
  const [tabValue, setTabValue] = useState(0);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue)
  }

  return (<>
    <Box sx={{ width: "100%", px: 3, py: 2, maxWidth: "25vw", borderRight: "1px solid #e0e0e0", }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2 }}>
        <Typography variant="h5" component="h1" sx={{ fontWeight: "bold" }}>
          Inbox
        </Typography>
        <Button
          variant="contained"
          startIcon={<Edit size={18} />}
          sx={{
            bgcolor: "black",
            color: "white",
            borderRadius: "6px",
            "&:hover": {
              bgcolor: "#333",
            },
          }}
        >
          Edit prompt
        </Button>
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
        <Tab label="All" />
        <Tab label="Unread" sx={{ color: "text.secondary" }} />
        <Tab label="Completed" sx={{ color: "text.secondary" }} />
        <Tab label="Discarded" sx={{ color: "text.secondary" }} />
      </Tabs>

      <List sx={{ p: 0 }}>
        <Paper
          elevation={0}
          sx={{
            border: "1px solid #e0e0e0",
            borderRadius: 2,
            mb: 2,
            overflow: "hidden",
          }}
        >
          <ListItem sx={{ py: 2, px: 3, bgcolor: "#f8f8f8" }}>
            <Stack direction="column" spacing={1} width="100%">
              <Stack direction="row" spacing={1} alignItems="center">
                <Box
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    color: "green",
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
                      bgcolor: "green",
                      mr: 0.5,
                    }}
                  />
                  100%
                </Box>
                <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                  •
                </Typography>
                <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                  r/sales
                </Typography>
                <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                  •
                </Typography>
                <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                  about 4 hours ago
                </Typography>
              </Stack>
              <Typography variant="body1" sx={{ fontWeight: "medium" }}>
                Sales to developing markets
              </Typography>
            </Stack>
          </ListItem>
        </Paper>

        <ListItem sx={{ py: 2, px: 3 }}>
          <Stack direction="column" spacing={1} width="100%">
            <Stack direction="row" spacing={1} alignItems="center">
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  color: "green",
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
                    bgcolor: "green",
                    mr: 0.5,
                  }}
                />
                100%
              </Box>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                r/sales
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                about 4 hours ago
              </Typography>
            </Stack>
            <Typography variant="body1" sx={{ fontWeight: "medium" }}>
              28 years as an Individual contributor, looking to move to become a sales manager
            </Typography>
          </Stack>
        </ListItem>
        <Divider />

        <ListItem sx={{ py: 2, px: 3 }}>
          <Stack direction="column" spacing={1} width="100%">
            <Stack direction="row" spacing={1} alignItems="center">
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  color: "green",
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
                    bgcolor: "green",
                    mr: 0.5,
                  }}
                />
                100%
              </Box>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                r/marketing
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                about 7 hours ago
              </Typography>
            </Stack>
            <Typography variant="body1" sx={{ fontWeight: "medium" }}>
              When you gain a new client, do you draw up the contract/agreement or them?
            </Typography>
          </Stack>
        </ListItem>
        <Divider />

        <ListItem sx={{ py: 2, px: 3 }}>
          <Stack direction="column" spacing={1} width="100%">
            <Stack direction="row" spacing={1} alignItems="center">
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  color: "green",
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
                    bgcolor: "green",
                    mr: 0.5,
                  }}
                />
                100%
              </Box>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                r/marketing
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                about 8 hours ago
              </Typography>
            </Stack>
            <Typography variant="body1" sx={{ fontWeight: "medium" }}>
              Where's the best place to get rack cards, business cards, etc printed?
            </Typography>
          </Stack>
        </ListItem>
        <Divider />

        <ListItem sx={{ py: 2, px: 3 }}>
          <Stack direction="column" spacing={1} width="100%">
            <Stack direction="row" spacing={1} alignItems="center">
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  color: "#9ACD32",
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
                    bgcolor: "#9ACD32",
                    mr: 0.5,
                  }}
                />
                80%
              </Box>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                r/marketing
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                about 11 hours ago
              </Typography>
            </Stack>
            <Typography variant="body1" sx={{ fontWeight: "medium" }}>
              Need career path help
            </Typography>
          </Stack>
        </ListItem>
        <Divider />

        <ListItem sx={{ py: 2, px: 3 }}>
          <Stack direction="column" spacing={1} width="100%">
            <Stack direction="row" spacing={1} alignItems="center">
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  color: "green",
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
                    bgcolor: "green",
                    mr: 0.5,
                  }}
                />
                100%
              </Box>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                r/sales
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", mx: 1 }}>
                •
              </Typography>
              <Typography component="span" sx={{ fontSize: "0.875rem", color: "text.secondary" }}>
                about 11 hours ago
              </Typography>
            </Stack>
            <Typography variant="body1" sx={{ fontWeight: "medium" }}>
              On a 9 month plan to become a sales manager with a new team from being an AE after 10...
            </Typography>
          </Stack>
        </ListItem>
      </List>
    </Box>
  </>);
}

export default InboxComponent;