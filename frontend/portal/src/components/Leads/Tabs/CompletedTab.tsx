"use client"

import React, { useEffect } from "react";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import ListRenderComp from "./LeadListComp";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";
import { LeadTabStatus, setActiveTab, setError, setIsLoading, setListOfLeads } from "../../../../store/Lead/leadSlice";

const CompletedTabComponent = () => {
    const { portalClient } = useClientsContext();
    const dispatch = useAppDispatch();
    const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

    useEffect(() => {

        const getAllLeadsByStatus = async () => {
            dispatch(setIsLoading(true));
            dispatch(setActiveTab(LeadTabStatus.COMPLETED));

            try {
                const result = await portalClient.getLeadsByStatus({ status: LeadStatus.COMPLETED });
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

    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [(selectedleadData === null)]);

    return (<ListRenderComp />);
};

export default CompletedTabComponent;
