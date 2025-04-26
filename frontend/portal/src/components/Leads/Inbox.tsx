"use client"

import React from "react";
import {
  Box,
  Typography,
  Tabs,
  Tab,
  Button,
} from "@mui/material";
import NewTabComponent from "./Tabs/NewTab";
import CompletedTabComponent from "./Tabs/CompletedTab";
import { useRedditIntegrationStatus } from "./Tabs/useRedditIntegrationStatus";
import { routes } from "@doota/ui-core/routing";
import Link from "next/link";
import DiscardedTabComponent from "./Tabs/DiscardedTab";
import { RedditLead } from "@doota/pb/doota/reddit/v1/reddit_pb";

export interface ChildComponentProps {
  selectedleadData: RedditLead | null;
  setSelectedLeadData: React.Dispatch<React.SetStateAction<RedditLead | null>>;
}

const InboxComponent: React.FC<ChildComponentProps> = ({ selectedleadData, setSelectedLeadData }) => {
  const [tabValue, setTabValue] = React.useState<number>(0);
  const { isConnected } = useRedditIntegrationStatus();

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  return (
    <Box
      sx={{
        width: "100%",
        py: 2,
        maxWidth: selectedleadData ? "25vw" : "100%",
        borderRight: "1px solid #e0e0e0",
      }}
    >
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          mt: 5,
          mx: 5,
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
          mx: 5,
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
        <Tab label="New" sx={{ color: "text.secondary" }} />
        <Tab label="Completed" sx={{ color: "text.secondary" }} />
        <Tab label="Discarded" sx={{ color: "text.secondary" }} />
      </Tabs>

      {isConnected ? (<>
        {tabValue === 0 && <NewTabComponent selectedleadData={selectedleadData} setSelectedLeadData={setSelectedLeadData} />}
        {tabValue === 1 && <CompletedTabComponent selectedleadData={selectedleadData} setSelectedLeadData={setSelectedLeadData} />}
        {tabValue === 2 && <DiscardedTabComponent selectedleadData={selectedleadData} setSelectedLeadData={setSelectedLeadData} />}
      </>) : (
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            height: "100%",
            textAlign: "center",
            px: 2,
          }}
        >
          <Typography variant="body1" color="text.secondary">
            {`Please connect your reddit account and a button. On clicking it, it should redirect me to settings/account.`}
          </Typography>
          <Button variant="contained" component={Link} href={routes.app.settings.account} sx={{ mt: 4 }}>
            {`Connect`}
          </Button>
        </Box>
      )}
    </Box>
  );
};

export default InboxComponent;
