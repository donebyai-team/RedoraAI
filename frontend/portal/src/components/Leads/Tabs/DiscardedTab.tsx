"use client"

import React, { useEffect } from "react";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";
import ListRenderComp from "./LeadListComp";
import { LeadTabStatus, setActiveTab, setError, setIsLoading, setListOfLeads } from "../../../../store/Lead/leadSlice";

const DiscardedTabComponent = () => {
    const { portalClient } = useClientsContext();
    const dispatch = useAppDispatch();
    const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

    useEffect(() => {

        const getAllLeadsByStatus = async () => {
            dispatch(setIsLoading(true));
            dispatch(setActiveTab(LeadTabStatus.DISCARDED));

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
