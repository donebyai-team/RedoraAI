'use client';

import React, {useEffect, useState} from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Wand2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {Source, SubscriptionPlanID} from "@doota/pb/doota/core/v1/core_pb";
import {useClientsContext} from "@doota/ui-core/context/ClientContext";
import {PostInsight} from "@doota/pb/doota/core/v1/insight_pb";

const toneOptions = [
    { value: "professional", label: "Professional" },
    { value: "casual", label: "Casual" },
    { value: "friendly", label: "Friendly" },
];

const goalOptions = [
    { value: "karma", label: "Build Karma" },
    { value: "feedback", label: "Get Feedback" },
    { value: "leads", label: "Generate Leads" },
];

export default function CreatePost() {
    const router = useRouter();
    const { portalClient } = useClientsContext();

    const [isFetching, setIsFetching] = useState<boolean>(false);
    const [insights, setInsights] = useState<PostInsight[]>([]);
    const [sources, setSources] = useState<Source[]>([]);

    const [selectedInsight, setSelectedInsight] = useState<string>("");
    const [selectedSubreddit, setSelectedSubreddit] = useState<string>("");
    const [customTopic, setCustomTopic] = useState("");
    const [postDetails, setPostDetails] = useState("");
    const [selectedGoal, setSelectedGoal] = useState("");
    const [selectedTone, setSelectedTone] = useState("");

    useEffect(() => {
       setIsFetching(true);
        portalClient
            .getInsights({})
            .then((res) => {
                setInsights(res.insights);
            })
            .catch((err) => {
                console.error("Error fetching insights:", err);
            })
            .finally(() => {
                setIsFetching(false);
            });

        portalClient.getSources({})
            .then((res) => {
                setSources(res.sources);
            })
            .catch((err) => {
                console.error("Error fetching insights:", err);
            })
            .finally(() => {
                setIsFetching(false);
            });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const handleGeneratePost = () => {
        const postData = {
            referenceId: selectedInsight,
            sourceId: selectedSubreddit,
            topic: customTopic,
            context: postDetails,
            goal: selectedGoal,
            tone: selectedTone,
        };

        localStorage.setItem("generatedPostData", JSON.stringify(postData));
        portalClient.createPost(postData).then((res) => {
            console.log("Post created successfully:", res);
        }).catch((err) => {
            console.error("Error creating post:", err);
        })
        router.push("/post-creation-hub/editor");
    };

    return (
        <div className="p-6 ml-[10%] mr-[10%]">
            <h1 className="text-2xl font-bold mb-1">Create New Post</h1>
            <p className="text-gray-500 mb-6">Generate AI-powered Reddit posts from insights or custom topics</p>

            <Card>
                <CardContent className="p-6 space-y-6">
                    {/* Suggested Topic */}
                    <div>
                        <Label className="mb-1 block">Suggested Topics from Insights (Optional)</Label>
                        <Select onValueChange={setSelectedInsight}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a suggested topic or leave blank to add your own..." />
                            </SelectTrigger>
                            <SelectContent>
                                {insights.map((insight) => (
                                    <SelectItem key={insight.id} value={insight.id}>
                                        <div className="flex items-center gap-2">
                                            <Badge variant="secondary" className="text-xs">
                                                {insight.relevancyScore}%
                                            </Badge>
                                            <span className="truncate max-w-[300px] text-sm">{insight.topic}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    {/* Topic */}
                    <div>
                        <Label htmlFor="topic" className="mb-1 block">Topic</Label>
                        <Textarea
                            id="topic"
                            value={customTopic}
                            onChange={(e) => setCustomTopic(e.target.value)}
                            placeholder="Enter your topic..."
                            className="min-h-[100px] text-base"
                        />
                    </div>

                    {/* Post Details */}
                    <div>
                        <Label htmlFor="details" className="mb-1 block">Post Details & Context</Label>
                        <Textarea
                            id="details"
                            value={postDetails}
                            onChange={(e) => setPostDetails(e.target.value)}
                            placeholder="Add specific details, context, examples, or requirements for your post..."
                            className="min-h-[150px] text-base"
                        />
                    </div>

                    {/* Dropdowns */}
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                        <div>
                            <Label className="mb-1 block">Target Subreddit</Label>
                            <Select value={selectedSubreddit} onValueChange={setSelectedSubreddit}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select subreddit" />
                                </SelectTrigger>
                                <SelectContent>
                                    {sources.map((source) => (
                                        <SelectItem key={source.id} value={source.id}>
                                            {source.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div>
                            <Label className="mb-1 block">Post Goal</Label>
                            <Select value={selectedGoal} onValueChange={setSelectedGoal}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select goal" />
                                </SelectTrigger>
                                <SelectContent>
                                    {goalOptions.map((goal) => (
                                        <SelectItem key={goal.value} value={goal.value}>
                                            {goal.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div>
                            <Label className="mb-1 block">Tone</Label>
                            <Select value={selectedTone} onValueChange={setSelectedTone}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select tone" />
                                </SelectTrigger>
                                <SelectContent>
                                    {toneOptions.map((tone) => (
                                        <SelectItem key={tone.value} value={tone.value}>
                                            {tone.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    {/* Button */}
                    <div className="flex justify-center pt-4">
                        <Button
                            onClick={handleGeneratePost}
                            disabled={!customTopic || !selectedSubreddit || !selectedGoal || !selectedTone}
                            className="px-8 text-base"
                        >
                            <Wand2 className="h-4 w-4 mr-2" />
                            Generate Post with AI
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
