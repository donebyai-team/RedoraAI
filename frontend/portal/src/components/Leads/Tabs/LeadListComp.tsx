"use client"

import React from "react";
import {
    Box,
    Typography,
    List,
    ListItem,
    Stack,
} from "@mui/material";
import { LeadTyeps, setSelectedLeadData } from "../../../../store/Lead/leadSlice";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { formateDate, getSubredditName, setLeadActive } from "./NewTab";
import { RootState } from "../../../../store/store";
import { LoadigSkeletons } from "../../NavBar";

interface ListRenderCompProps {
    list: LeadTyeps[];
    isLoading: boolean
}
const ListRenderComp: React.FC<ListRenderCompProps> = ({ isLoading, list }) => {

    const dispatch = useAppDispatch();
    const { subredditList } = useAppSelector((state: RootState) => state.source);
    const { selectedleadData } = useAppSelector((state: RootState) => state.lead);

    const handleSelectedLead = (data: LeadTyeps) => {
        dispatch(setSelectedLeadData(data));
    };

    return (
        isLoading ?
            <Box sx={{ display: 'flex', px: 4, flexDirection: "column", alignItems: "center", height: "100%", width: "100%", gap: 2, mt: 5 }}>
                <LoadigSkeletons count={5} height={60} />
            </Box>
            :
            <Box sx={{ width: "100%", pt: 2, height: "83dvh", overflowY: "scroll" }}>
                {(list.length > 0) ? (
                    <List sx={{ p: 0, mx: 4 }}>
                        {list.map((post, index) => (
                            <React.Fragment key={index}>
                                <ListItem onClick={() => handleSelectedLead(post)} sx={{ p: 3, mb: (index !== list.length - 1) ? 2 : 0, cursor: "pointer", ...setLeadActive(selectedleadData?.id as string, post.id) }}>
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
                                                {getSubredditName(subredditList, post.sourceId)}
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
                            </React.Fragment>
                        ))}
                    </List>
                ) : (
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
                            {`Sit back and relax, we are finding relevant leads for you. We will notify you once it’s ready.`}
                        </Typography>
                    </Box>
                )}
            </Box>
    );
};

export default ListRenderComp;
