"use client"

import React, { useEffect, useState } from "react";
import {
    Box,
    Typography,
    Tabs,
    Tab,
    List,
    ListItem,
    Divider,
    Stack,
    CircularProgress,
} from "@mui/material";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import toast from "react-hot-toast";
import { RedditLead, SubReddit } from "@doota/pb/doota/reddit/v1/reddit_pb";
import { formatDistanceToNow } from 'date-fns';
import { useSearchParams } from "next/navigation";
import { Timestamp } from "@bufbuild/protobuf/wkt";

export const formateDate = (timestamp: Timestamp): string => {
    const millis = Number(timestamp.seconds) * 1000; // convert bigint to number
    const date = new Date(millis);
    const timeAgo = formatDistanceToNow(date, { addSuffix: true });
    return timeAgo;
};

export const getSubredditName = (list: SubReddit[], id: string) => {
    const name = list?.find(reddit => reddit.id === id)?.name ?? "N/A";
    return name;
};

const NewTabComponent = () => {
    const { portalClient } = useClientsContext();
    const searchParams = useSearchParams()
    const relevancyScoreParam = searchParams.get('relevancy_score');
    const relevancyScore = relevancyScoreParam && !isNaN(Number(relevancyScoreParam)) ? Number(relevancyScoreParam) : "";
    const subReddit = searchParams.get('currentActiveSubRedditId') ?? "";
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [listofleads, setListOfLeads] = useState<RedditLead[]>([]);
    const [subredditList, setSubredditList] = useState<SubReddit[]>([]);

    useEffect(() => {

        const getAllRelevantLeads = async () => {
            setIsLoading(true);

            try {
                const result = await portalClient.getRelevantLeads({ ...(relevancyScore && { relevancyScore }), ...(subReddit && { subReddit }) });
                setListOfLeads(result?.leads ?? []);
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong"
                toast.error(message);
            } finally {
                setIsLoading(false);
            }
        }
        getAllRelevantLeads();

    }, [relevancyScore, subReddit]);

    useEffect(() => {

        const getAllSubReddits = async () => {

            try {
                const result = await portalClient.getSubReddits({});
                setSubredditList(result?.subreddits ?? []);
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong"
                console.log(message);
            }
        }
        getAllSubReddits();

    }, []);

    return (
        isLoading ?
            <Box sx={{ display: 'flex', flexDirection: "column", alignItems: "center", height: "100vh", width: "100%" }}>
                <CircularProgress />
            </Box>
            :
            <Box sx={{ width: "100%", px: 3, py: 2 }}>
                {listofleads.length === 0 ? (
                    <Box
                        sx={{
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            height: "100vh",
                            textAlign: "center",
                            px: 2,
                        }}
                    >
                        <Typography variant="body1" color="text.secondary">
                            {`Sit back and relax, we are finding relevant leads for you. We will
                            notify you once it’s ready.`}
                        </Typography>
                    </Box>
                ) : (
                    <List sx={{ p: 0 }}>
                        {listofleads.map((post, index) => (
                            <React.Fragment key={index}>
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
                                                        mr: 1,
                                                    }}
                                                />
                                                {post.relevancyScore}%
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
                                                {getSubredditName(subredditList, post.subredditId)}
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
                                                {post.postCreatedAt ? formateDate(post.postCreatedAt) : "N/A"}
                                            </Typography>
                                        </Stack>
                                        <Typography variant="body1" sx={{ fontWeight: "medium" }}>
                                            {post.title}
                                        </Typography>
                                    </Stack>
                                </ListItem>
                                {index !== listofleads.length - 1 && <Divider />}
                            </React.Fragment>
                        ))}
                    </List>
                )}
            </Box>
    );
};

export default NewTabComponent;
