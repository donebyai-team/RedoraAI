"use client"

import React, { useEffect, useState } from "react";
import {
    Box,
    Typography,
    List,
    ListItem,
    Divider,
    Stack,
    CircularProgress,
} from "@mui/material";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { Lead, LeadStatus, Source } from "@doota/pb/doota/core/v1/core_pb";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";
import ListRenderComp from "./LeadListComp";
import { setError, setIsLoading, setListOfLeads } from "../../../../store/Lead/leadSlice";

const DiscardedTabComponent = () => {
    const { portalClient } = useClientsContext();
    const dispatch = useAppDispatch();
    const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

    useEffect(() => {

        const getAllLeadsByStatus = async () => {
            dispatch(setIsLoading(true));

            try {
                const result = await portalClient.getLeadsByStatus({ status: LeadStatus.NOT_RELEVANT });
                dispatch(setListOfLeads(result?.leads ?? []));
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong"
                toast.error(message);
                dispatch(setError(message));
            } finally {
                dispatch(setIsLoading(false));
            }
        }
        getAllLeadsByStatus();

    }, [(selectedleadData === null)]);

    return (<ListRenderComp />);
};

export default DiscardedTabComponent;
