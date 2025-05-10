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
import { formateDate, isSameDay, setLeadActive } from "./NewTab";
import { RootState } from "../../../../store/store";
import { LoadigSkeletons } from "../../NavBar";
import { MarkdownRenderer } from "../../Html/HtmlRenderer";

interface ListRenderCompProps {
    list: LeadTyeps[];
    isLoading: boolean
}

const ListRenderComp: React.FC<ListRenderCompProps> = ({ isLoading, list }) => {

    const dispatch = useAppDispatch();
    // const { subredditList } = useAppSelector((state: RootState) => state.source);
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
                {(list?.length > 0) ? (
                    <List sx={{ p: 0, mx: 4 }}>
                        {list.map((post, index) => (
                            <React.Fragment key={index}>
                                <ListItem onClick={() => handleSelectedLead(post)} sx={{ p: 3, mb: (index !== list.length - 1) ? 2 : 0, cursor: "pointer", ...setLeadActive(selectedleadData?.id as string, post.id) }}>
                                    <Stack direction="column" spacing={1} width="100%">
                                        <Stack
                                            direction="row"
                                            spacing={1}
                                            alignItems="center"
                                            flexWrap="wrap"
                                            useFlexGap
                                        >
                                            <Box
                                                sx={{
                                                    display: "flex",
                                                    alignItems: "center",
                                                    color: "green",
                                                    fontSize: "0.875rem",
                                                    wordBreak: "break-word",
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
                                                        flexShrink: 0,
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
                                                    wordBreak: "break-word",
                                                }}
                                            >
                                                {post.metadata?.subredditPrefixed}
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
                                                    wordBreak: "break-word",
                                                }}
                                            >
                                                Post: {post.postCreatedAt ? formateDate(post.postCreatedAt) : "N/A"}
                                            </Typography>
                                        </Stack>

                                        <Typography variant="body1" sx={{ fontWeight: "medium" }}>
                                            <MarkdownRenderer data={post.title ?? ""} />
                                        </Typography>

                                        {/* Scraped On line */}
                                        <Typography sx={{ fontSize: "0.7rem", color: "text.secondary" }}>
                                            Match: {post.createdAt ? formateDate(post.createdAt) : "N/A"}
                                        </Typography>

                                        <Box
                                            sx={{
                                                display: "inline-block",
                                                backgroundColor: "#4CAF50", // Lighter green
                                                color: "white",
                                                fontSize: "0.7rem", // Slightly larger text
                                                px: 1.2, // More horizontal padding
                                                py: 0.3, // More vertical padding
                                                borderRadius: "6px", // Slightly more rounded corners
                                                width: "fit-content",
                                                mt: 0.7, // More spacing from "Scraped On"
                                                fontWeight: "bold", // Make text stand out a little more
                                            }}
                                        >
                                            Keyword: {post.keyword?.name}
                                        </Box>
                                        {post.createdAt && isSameDay(post.createdAt) && (
                                            <Box
                                                sx={{
                                                    display: "flex",
                                                    flexWrap: "wrap",
                                                    gap: 1,
                                                    mt: 0.5,
                                                }}
                                            >
                                                <Box
                                                    sx={{
                                                        backgroundColor: '#FFCDD2', // light red background
                                                        color: '#B71C1C',           // dark red text
                                                        fontSize: "0.8rem",
                                                        px: 1.5,
                                                        py: 0.5,
                                                        mt: "0.5rem",
                                                        borderRadius: "999px",
                                                        whiteSpace: "nowrap",
                                                        fontWeight: 500,
                                                    }}
                                                >
                                                    🔥 New
                                                </Box>
                                            </Box>
                                        )}



                                        {/* {post.intents && post.intents.length > 0 && (
                                            <Box
                                                sx={{
                                                    display: "flex",
                                                    flexWrap: "wrap",
                                                    gap: 1,
                                                    mt: 0.5,
                                                }}
                                            >
                                                {post.intents.map((label: string, idx: number) => {
                                                    let emoji = '';
                                                    let bgColor = ''; // Default background color
                                                    let textColor = '';

                                                    if (label === 'PROBABLE_LEAD') {
                                                        emoji = '🔥 '; // Emoji for Probable Lead
                                                        bgColor = '#F8BBD0';
                                                        textColor = '#880E4F';
                                                    } else if (label === 'BEST_FOR_ENGAGEMENT') {
                                                        emoji = '🌟 '; // Emoji for Best for Engagement
                                                        bgColor = '#D1C4E9';
                                                        textColor = '#4A148C';
                                                    }

                                                    return (
                                                        <Box
                                                            key={idx}
                                                            sx={{
                                                                backgroundColor: bgColor,
                                                                color: textColor,
                                                                fontSize: "0.8rem",
                                                                px: 1.5,
                                                                py: 0.5,
                                                                mt: "0.5rem",
                                                                borderRadius: "999px", // full pill shape
                                                                whiteSpace: "nowrap",
                                                                fontWeight: 500,
                                                            }}
                                                        >
                                                            {emoji}{label}
                                                        </Box>
                                                    );
                                                })}
                                            </Box>
                                        )} */}
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
