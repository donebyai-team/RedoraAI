'use client';
import {
    Calendar, Save, Send, Eye, Wand2, Edit3, RefreshCw,
    TrendingUp, MessageSquare, Heart, AlertCircle, CheckCircle,
    Clock, X, ArrowLeft, Undo, History
} from "lucide-react";
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow
} from "@/components/ui/table";
import React from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {routes} from "@doota/ui-core/routing";

interface ScheduledPost {
    id: string;
    title: string;
    content: string;
    subreddit: string;
    status: "draft" | "scheduled" | "posted" | "failed";
    createdDate: string;
    scheduledDate?: string;
    postedDate?: string;
    engagement?: {
        upvotes: number;
        comments: number;
    };
    failureReason?: string;
}

const samplePosts: ScheduledPost[] = [
    {
        id: "1",
        title: "My experience with AI-generated solutions for enterprise",
        content: "After working with various AI tools for enterprise solutions...",
        subreddit: "SaaS",
        status: "posted",
        createdDate: "2025-07-01T09:00:00Z",
        scheduledDate: "2025-07-02T10:00:00Z",
        postedDate: "2025-07-02T10:00:00Z",
        engagement: { upvotes: 142, comments: 23 }
    },
    {
        id: "2",
        title: "Why most AI solutions fail at customization",
        content: "Here's what I've learned about AI limitations in enterprise...",
        subreddit: "startups",
        status: "scheduled",
        createdDate: "2025-07-02T14:00:00Z",
        scheduledDate: "2025-07-04T14:00:00Z"
    },
    {
        id: "3",
        title: "Building vs buying: The AI perspective",
        content: "When should you use AI vs custom development?",
        subreddit: "Entrepreneur",
        status: "failed",
        createdDate: "2025-07-01T12:00:00Z",
        scheduledDate: "2025-07-01T16:00:00Z",
        failureReason: "Reddit API rate limit exceeded. Post will be retried automatically."
    },
    {
        id: "4",
        title: "Draft: AI implementation strategies",
        content: "Exploring different approaches to AI implementation...",
        subreddit: "SaaS",
        status: "draft",
        createdDate: "2025-07-03T08:00:00Z"
    }
];

export default function Posts() {
    const router = useRouter();

    const getStatusIcon = (status: string) => {
        switch (status) {
            case "posted":
                return <CheckCircle className="h-4 w-4 text-green-600" />;
            case "scheduled":
                return <Clock className="h-4 w-4 text-blue-600" />;
            case "failed":
                return <AlertCircle className="h-4 w-4 text-red-600" />;
            case "draft":
                return <Edit3 className="h-4 w-4 text-gray-600" />;
            default:
                return null;
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case "posted":
                return "bg-green-100 text-green-800";
            case "scheduled":
                return "bg-blue-100 text-blue-800";
            case "failed":
                return "bg-red-100 text-red-800";
            case "draft":
                return "bg-gray-100 text-gray-800";
            default:
                return "bg-gray-100 text-gray-800";
        }
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit"
        });
    };

    const handleEditPost = (postId: string) => {
        const post = samplePosts.find((p) => p.id === postId);
        if (!post) return;

        const generatedPostData = {
            title: post.title,
            content: post.content,
            subreddit: post.subreddit,
            goal: "N/A", // You can update if goal info is available
            tone: "N/A", // You can update if tone info is available
        };

        localStorage.setItem("generatedPostData", JSON.stringify(generatedPostData));
        router.push(routes.new.postCreationHub.editor);
    };

    return (
        <div className="p-6">
            <div className="mb-6">
                <h1 className="text-2xl font-bold">Posts Management</h1>
                <p className="text-gray-600">
                    Manage your drafts, scheduled posts, and view engagement
                </p>
            </div>

            <Card>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Status</TableHead>
                                <TableHead>Title</TableHead>
                                <TableHead>Subreddit</TableHead>
                                <TableHead>Created</TableHead>
                                <TableHead>Scheduled/Posted</TableHead>
                                <TableHead>Engagement</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {samplePosts.map((post) => (
                                <TableRow key={post.id}>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            {getStatusIcon(post.status)}
                                            <Badge className={getStatusColor(post.status)}>{post.status}</Badge>
                                        </div>
                                    </TableCell>
                                    <TableCell className="max-w-xs">
                                        <div>
                                            <p className="font-medium truncate">{post.title}</p>
                                            {post.status === "failed" && post.failureReason && (
                                                <p className="text-xs text-red-600 mt-1">{post.failureReason}</p>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>r/{post.subreddit}</TableCell>
                                    <TableCell className="text-sm">{formatDate(post.createdDate)}</TableCell>
                                    <TableCell>
                                        <div className="text-sm">
                                            {post.status === "posted" && post.postedDate ? (
                                                <span>Posted: {formatDate(post.postedDate)}</span>
                                            ) : post.status === "scheduled" && post.scheduledDate ? (
                                                <span>Scheduled: {formatDate(post.scheduledDate)}</span>
                                            ) : (
                                                <span className="text-gray-400">-</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        {post.engagement ? (
                                            <div className="flex gap-3 text-sm">
                                                <div className="flex items-center gap-1">
                                                    <Heart className="h-3 w-3" />
                                                    {post.engagement.upvotes}
                                                </div>
                                                <div className="flex items-center gap-1">
                                                    <MessageSquare className="h-3 w-3" />
                                                    {post.engagement.comments}
                                                </div>
                                            </div>
                                        ) : (
                                            <span className="text-gray-400 text-sm">-</span>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex gap-1">
                                            <Button variant="outline" size="sm">
                                                <Eye className="h-3 w-3" />
                                            </Button>
                                            {post.status === "failed" && (
                                                <Button variant="outline" size="sm">
                                                    <RefreshCw className="h-3 w-3" />
                                                </Button>
                                            )}
                                            {(post.status === "scheduled" || post.status === "draft") && (
                                                <Button variant="outline" size="sm" onClick={() => handleEditPost(post.id)}>
                                                    <Edit3 className="h-3 w-3" />
                                                </Button>
                                            )}
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
