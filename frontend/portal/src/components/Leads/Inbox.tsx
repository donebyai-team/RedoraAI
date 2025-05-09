"use client"

import React from "react";
import {
  Box,
  Typography,
  Tabs,
  Tab,
  Button,
  Skeleton,
} from "@mui/material";
import NewTabComponent from "./Tabs/NewTab";
import CompletedTabComponent from "./Tabs/CompletedTab";
import { routes } from "@doota/ui-core/routing";
import Link from "next/link";
import DiscardedTabComponent from "./Tabs/DiscardedTab";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { RootState } from "../../../store/store";
import { LeadTabStatus, setActiveTab } from "../../../store/Lead/leadSlice";

const tabList: { label: string; value: LeadTabStatus }[] = [
  { label: "New", value: LeadTabStatus.NEW },
  { label: "Responded", value: LeadTabStatus.COMPLETED },
  { label: "Discarded", value: LeadTabStatus.DISCARDED },
];

const InboxComponent = () => {
  const dispatch = useAppDispatch();
  const { isConnected, loading: isLoading } = useAppSelector((state: RootState) => state.redditIntegration);
  const { activeTab, selectedleadData } = useAppSelector((state: RootState) => state.lead);

  const handleTabChange = (_: React.SyntheticEvent, newValue: LeadTabStatus) => {
    dispatch(setActiveTab(newValue));
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case LeadTabStatus.NEW:
        return <NewTabComponent />;
      case LeadTabStatus.COMPLETED:
        return <CompletedTabComponent />;
      case LeadTabStatus.DISCARDED:
        return <DiscardedTabComponent />;
      default:
        return null;
    }
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
        value={activeTab}
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
        {tabList.map((tab) => (
          <Tab
            key={tab.value}
            label={tab.label}
            value={tab.value}
            sx={{ color: "text.secondary" }}
          />
        ))}
      </Tabs>

      {isLoading ?
        <Box sx={{ display: 'flex', px: 4, flexDirection: "column", alignItems: "center", height: "100%", width: "100%", gap: 2, mt: 5 }}>
          {Array.from({ length: 5 }).map((_: any, index: number) => (
            <Skeleton key={index} variant="rounded" width={"100%"} height={60} />
          ))}
        </Box>
        :
        isConnected ? (
          renderTabContent()
        ) : (
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
              {`Please connect your Reddit account to continue. Click the button below to go to your account settings.`}
            </Typography>
            <Button
              variant="contained"
              component={Link}
              href={routes.app.settings.account}
              sx={{ mt: 4 }}
            >
              Connect
            </Button>
          </Box>
        )
      }

    </Box>
  );
};

export default InboxComponent;
