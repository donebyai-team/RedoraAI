'use client';
import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Calendar, History, Save, Undo } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

interface GenerationHistory {
    id: string;
    title: string;
    content: string;
    timestamp: string;
}

interface GeneratedPostData {
    title: string;
    content: string;
    subreddit: string;
    goal: string;
    tone: string;
}

export default function PostEditor() {
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [scheduledDate, setScheduledDate] = useState("");
    const [generationHistory, setGenerationHistory] = useState<GenerationHistory[]>([]);
    const [showHistory, setShowHistory] = useState(false);
    const [settings, setSettings] = useState<GeneratedPostData | null>(null);

    useEffect(() => {
        const stored = localStorage.getItem("generatedPostData");
        if (stored) {
            const parsed: GeneratedPostData = JSON.parse(stored);
            setTitle(parsed.title);
            setContent(parsed.content);
            setSettings(parsed);

            const initialGeneration: GenerationHistory = {
                id: "1",
                title: parsed.title,
                content: parsed.content,
                timestamp: new Date().toISOString(),
            };
            setGenerationHistory([initialGeneration]);
        }
    }, []);

    const handleRegenerate = () => {
        const newTitle = `Updated: ${title.split(':')[1] || title}`;
        const newContent = `Here's a fresh perspective...\n\nThis regenerated version takes a new approach.\n\n${content.split('\n\n')[1] || ''}`;

        const newGeneration: GenerationHistory = {
            id: (generationHistory.length + 1).toString(),
            title: newTitle,
            content: newContent,
            timestamp: new Date().toISOString(),
        };

        setGenerationHistory((prev) => [...prev, newGeneration]);
        setTitle(newTitle);
        setContent(newContent);
    };

    const handleSelectFromHistory = (item: GenerationHistory) => {
        setTitle(item.title);
        setContent(item.content);
        setShowHistory(false);
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
        new Date(timestamp).toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit'
        });

    return (
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
                                        {generationHistory.map((item, index) => (
                                            <div
                                                key={item.id}
                                                className="flex items-center justify-between p-3 bg-white rounded border cursor-pointer hover:bg-gray-50"
                                                onClick={() => handleSelectFromHistory(item)}
                                            >
                                                <div className="flex-1 min-w-0">
                                                    <p className="font-medium text-sm truncate">{item.title}</p>
                                                    <p className="text-xs text-gray-500 truncate">
                                                        {item.content.substring(0, 80)}...
                                                    </p>
                                                </div>
                                                <div className="flex items-center gap-2 ml-3">
                                                    <Badge variant="secondary" className="text-xs">
                                                        v{generationHistory.length - index}
                                                    </Badge>
                                                    <span className="text-xs text-gray-400">
                            {formatHistoryTime(item.timestamp)}
                          </span>
                                                </div>
                                            </div>
                                        ))}
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
                                <Button variant="outline" onClick={handleRegenerate}>
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
                                <Button onClick={handleSaveDraft} variant="outline" className="w-full">
                                    <Save className="h-4 w-4 mr-2" />
                                    Save as Draft
                                </Button>

                                <Button onClick={handleSchedule} className="w-full">
                                    <Calendar className="h-4 w-4 mr-2" />
                                    Schedule Post
                                </Button>
                            </div>

                            {settings && (
                                <div className="pt-4 border-t">
                                    <h4 className="font-medium mb-2">Post Settings</h4>
                                    <div className="space-y-2 text-sm text-gray-600">
                                        <p><strong>Subreddit:</strong> r/{settings.subreddit}</p>
                                        <p><strong>Goal:</strong> {settings.goal}</p>
                                        <p><strong>Tone:</strong> {settings.tone}</p>
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}