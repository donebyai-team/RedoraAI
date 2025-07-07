'use client';

import React, { useEffect, useState } from "react";
import {
    Card, CardContent, CardHeader, CardTitle
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
    Calendar, History, Loader2, Save, Undo
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useAppSelector } from "@/store/hooks";
import { PostRegenerationHistory, Post } from "@doota/pb/doota/core/v1/post_pb";
import { useCreatePost } from "@/components/hooks/useCreatePost";

interface ParsedHistory {
    title: string;
    description: string;
}

export default function PostEditor() {
    const { post } = useAppSelector((state) => state.postCreation);
    const { createPost } = useCreatePost();

    const [isLoading, setIsLoading] = useState(false);
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [scheduledDate, setScheduledDate] = useState("");
    const [generationHistory, setGenerationHistory] = useState<PostRegenerationHistory[]>([]);
    const [showHistory, setShowHistory] = useState(false);

    const [subreddit, setSubreddit] = useState("");
    const [goal, setGoal] = useState("");
    const [tone, setTone] = useState("");

    useEffect(() => {
        if (post) {
            setTitle(post.topic || "");
            setContent(post.description || "");
            setSubreddit(post.metadata?.settings?.sourceId || "");
            setGoal(post.metadata?.settings?.goal || "");
            setTone(post.metadata?.settings?.tone || "");
            setGenerationHistory(post.metadata?.history || []);
        }
    }, [post]);

    const handleRegenerate = async () => {
        if (!post) return;

        const postData = {
            id: post.id,
            sourceId: post.source,
            topic: post.metadata?.settings?.topic || "",
            context: post.metadata?.settings?.context || "",
            goal: goal,
            tone: tone
        };

        setIsLoading(true);
        const res: Post | undefined = await createPost(postData, false, setIsLoading);
        setIsLoading(false);

        if (res) {
            setTitle(res.topic || "");
            setContent(res.description || "");
            setGenerationHistory(res.metadata?.history || []);
        }
    };

    const handleSelectFromHistory = (item: PostRegenerationHistory) => {
        try {
            setTitle(item?.title || "");
            setContent(item?.description || "");
            setShowHistory(false);
        } catch (error) {
            console.error("Failed to parse generation history text:", error);
        }
    };

    const handleSaveDraft = () => {
        console.log("Saving as draft...");
    };

    const handleSchedule = () => {
        if (!scheduledDate) {
            alert("Please select a schedule date.");
            return;
        }
        console.log("Scheduling post for:", scheduledDate);
    };

    const formatHistoryTime = (timestamp: string) =>
        new Date(timestamp).toLocaleTimeString("en-US", {
            hour: "2-digit",
            minute: "2-digit",
        });

    return (
        <div>
            {isLoading ? (
                <div className="flex justify-center items-center my-14">
                    <Loader2 className="animate-spin" size={35} />
                </div>
            ) : (
                <div className="p-6 max-w-4xl mx-auto">
                    <div className="mb-6 flex items-center gap-4">
                        <div>
                            <h1 className="text-2xl font-bold">Edit Generated Post</h1>
                            <p className="text-gray-600">Review, edit, and schedule your AI-generated posts</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                        <div className="lg:col-span-2">
                            <Card>
                                <CardHeader>
                                    <div className="flex items-center justify-between">
                                        <CardTitle>Post Editor</CardTitle>
                                        {generationHistory.length > 0 && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => setShowHistory(!showHistory)}
                                            >
                                                <History className="h-4 w-4 mr-2" />
                                                History ({generationHistory.length})
                                            </Button>
                                        )}
                                    </div>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    {showHistory && generationHistory.length > 0 && (
                                        <div className="border rounded-lg p-4 bg-gray-50">
                                            <h4 className="font-medium mb-3">Generation History</h4>
                                            <div className="space-y-2 max-h-60 overflow-y-auto">
                                                {[...generationHistory].reverse().map((item, index) => {
                                                    return (
                                                        <div
                                                            key={index}
                                                            className="flex items-center justify-between p-3 bg-white rounded border cursor-pointer hover:bg-gray-50"
                                                            onClick={() => handleSelectFromHistory(item)}
                                                        >
                                                            <div className="flex-1 min-w-0">
                                                                <p className="font-medium text-sm truncate">{item.title}</p>
                                                                <p className="text-xs text-gray-500 truncate">
                                                                    {item.description?.substring(0, 88)}...
                                                                </p>
                                                            </div>
                                                            <div className="flex items-center gap-2 ml-3">
                                                                <Badge variant="secondary" className="text-xs">
                                                                    v{generationHistory.length - index}
                                                                </Badge>
                                                            </div>
                                                        </div>
                                                    );
                                                })}
                                            </div>
                                        </div>
                                    )}

                                    <div>
                                        <Label htmlFor="posts-title">Title</Label>
                                        <Input
                                            id="posts-title"
                                            value={title}
                                            onChange={(e) => setTitle(e.target.value)}
                                            placeholder="Enter post title..."
                                            className="text-base"
                                        />
                                    </div>

                                    <div>
                                        <Label htmlFor="posts-content">Content</Label>
                                        <Textarea
                                            id="posts-content"
                                            value={content}
                                            onChange={(e) => setContent(e.target.value)}
                                            placeholder="Write your post content..."
                                            className="min-h-[400px] text-base"
                                        />
                                    </div>

                                    <div className="flex gap-2">
                                        <Button
                                            variant="outline"
                                            onClick={handleRegenerate}
                                            disabled={isLoading || !post?.topic}
                                        >
                                            <Undo className="h-4 w-4 mr-2" />
                                            Regenerate
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>

                        <div>
                            <Card>
                                <CardHeader>
                                    <CardTitle>Post Actions</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div>
                                        <Label htmlFor="schedule-date">Schedule Date & Time</Label>
                                        <Input
                                            id="schedule-date"
                                            type="datetime-local"
                                            value={scheduledDate}
                                            onChange={(e) => setScheduledDate(e.target.value)}
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Button
                                            onClick={handleSaveDraft}
                                            variant="outline"
                                            className="w-full"
                                            disabled={isLoading || !post?.topic}
                                        >
                                            <Save className="h-4 w-4 mr-2" />
                                            Save as Draft
                                        </Button>

                                        <Button
                                            onClick={handleSchedule}
                                            className="w-full"
                                            disabled={isLoading || !post?.topic}
                                        >
                                            <Calendar className="h-4 w-4 mr-2" />
                                            Schedule Post
                                        </Button>
                                    </div>

                                    {(subreddit || goal || tone) && (
                                        <div className="pt-4 border-t">
                                            <h4 className="font-medium mb-2">Post Settings</h4>
                                            <div className="space-y-2 text-sm text-gray-600">
                                                {subreddit && <p><strong>Subreddit:</strong> r/{subreddit}</p>}
                                                {goal && <p><strong>Goal:</strong> {goal}</p>}
                                                {tone && <p><strong>Tone:</strong> {tone}</p>}
                                            </div>
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}