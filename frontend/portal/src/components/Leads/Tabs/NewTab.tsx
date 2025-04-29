"use client"

import React, { useEffect } from "react";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { formatDistanceToNow } from 'date-fns';
import { useSearchParams } from "next/navigation";
import { Timestamp } from "@bufbuild/protobuf/wkt";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { RootState } from "../../../../store/store";
import { LeadTabStatus, setActiveTab, setError, setIsLoading, setListOfLeads, setSelectedLeadData } from "../../../../store/Lead/leadSlice";
import ListRenderComp from "./LeadListComp";
import { SourceTyeps } from "../../../../store/Source/sourceSlice";

export const formateDate = (timestamp: Timestamp): string => {
    const millis = Number(timestamp.seconds) * 1000; // convert bigint to number
    const date = new Date(millis);
    const timeAgo = formatDistanceToNow(date, { addSuffix: true });
    return timeAgo;
};

export const getSubredditName = (list: SourceTyeps[], id: string) => {
    const name = list?.find(reddit => reddit.id === id)?.name ?? "N/A";
    return name;
};

export const setLeadActive = (parems_id: string, id: string) => {
    if (parems_id === id) {
        return ({ border: "1px solid #000", backgroundColor: "#0f172a0d", borderRadius: "0.5rem" });
    }
    return ({});
};

const NewTabComponent = () => {
    const { portalClient } = useClientsContext();
    const searchParams = useSearchParams()
    const relevancyScoreParam = searchParams.get('relevancy_score');
    const relevancyScore = relevancyScoreParam && !isNaN(Number(relevancyScoreParam)) ? Number(relevancyScoreParam) : "";
    const subReddit = searchParams.get('currentActiveSubRedditId') ?? "";
    const dispatch = useAppDispatch();
    const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

    useEffect(() => {

        const getAllRelevantLeads = async () => {
            dispatch(setIsLoading(true));
            dispatch(setActiveTab(LeadTabStatus.NEW));

            try {
                const result = await portalClient.getRelevantLeads({ ...(relevancyScore && { relevancyScore }), ...(subReddit && { subReddit }) });
                dispatch(setListOfLeads(result?.leads ?? []));
                dispatch(setSelectedLeadData(result?.leads[0]));
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong"
                toast.error(message);
                dispatch(setError(message));
            } finally {
                dispatch(setIsLoading(false));
            }
        }
        getAllRelevantLeads();

    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [relevancyScore, subReddit, (selectedleadData === null)]);

    return (<ListRenderComp />);
};

export default NewTabComponent;
