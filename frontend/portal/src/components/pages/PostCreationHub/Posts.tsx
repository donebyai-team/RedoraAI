'use client';
import {
    Calendar, Save, Send, Eye, Wand2, Edit3, RefreshCw,
    TrendingUp, MessageSquare, Heart, AlertCircle, CheckCircle,
    Clock, X, ArrowLeft, Undo, History, Loader2, Trash2
} from "lucide-react";
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow
} from "@/components/ui/table";
import React, {useEffect, useState} from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {routes} from "@doota/ui-core/routing";
import { AugmentedPost, Post} from "@doota/pb/doota/core/v1/post_pb";
import {portalClient} from "@/services/grpc";
import {useAppDispatch} from "@/store/hooks";
import { setPost } from "@/store/PostCreation/PostCreationSlice";

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

export enum PostStatus {
    CREATED = "CREATED",
    PROCESSING = "PROCESSING",
    SENT = "SENT",
    FAILED = "FAILED",
    SCHEDULED = "SCHEDULED",
}

export default function Posts() {
    const router = useRouter();
    const dispatch = useAppDispatch();
    const [isLoading, setIsLoading] = useState(false)
    const [posts, setPosts] = useState<AugmentedPost[]>([])

    useEffect(() => {
        setIsLoading(true)
        Promise.all([
            portalClient.getPosts({}).then(res => setPosts(res.posts)),
        ])
            .catch(err => console.error('Error fetching data:', err))
            .finally(() => setIsLoading(false))
    }, [portalClient])

    const getStatusIcon = (status?: string) => {
        switch (status) {
            case PostStatus.SENT:
                return <CheckCircle className="h-4 w-4 text-green-600" />;
            case PostStatus.SCHEDULED:
                return <Clock className="h-4 w-4 text-blue-600" />;
            case PostStatus.FAILED:
                return <AlertCircle className="h-4 w-4 text-red-600" />;
            case PostStatus.CREATED:
                return <Edit3 className="h-4 w-4 text-gray-600" />;
            case PostStatus.PROCESSING:
                return <Edit3 className="h-4 w-4 text-gray-600" />;
            default:
                return null;
        }
    };

    const getStatusColor = (status?: string) => {
        switch (status) {
            case PostStatus.SENT:
                return "bg-green-100 text-green-800";
            case PostStatus.SCHEDULED:
                return "bg-blue-100 text-blue-800";
            case PostStatus.FAILED:
                return "bg-red-100 text-red-800";
            case PostStatus.CREATED:
                return "bg-gray-100 text-gray-800";
            default:
                return "bg-gray-100 text-gray-800";
        }
    };

    const formatProtoTimestampUTC = (
        timestamp?: { seconds: bigint; nanos: number }
    ): string => {
        if (!timestamp) return "";

        const millis = Number(timestamp.seconds) * 1000 + Math.floor(timestamp.nanos / 1_000_000);
        const date = new Date(millis);

        return date.toLocaleString("en-US", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            timeZone: "UTC",
            hour12: false,
        });
    };


    const handleEditPost = (post: Post | undefined) => {
        if(post) {
            dispatch(setPost(post));
            router.push(routes.new.postCreationHub.editor);
        }
    };

    const handleDeletePost = (post: Post | undefined) => {
        console.log('Delete post:', post);
    }

    return (
        <>
            {
                isLoading ? (
                        <div className='flex justify-center items-center my-14'>
                            <Loader2 className='animate-spin' size={35} />
                        </div>
                    )
                    : (
                    <div className="p-6">
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
                                            {/*<TableHead>Engagement</TableHead>*/}
                                            <TableHead>Actions</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {posts.map((post) => (
                                            <TableRow key={post.post?.id}>
                                                <TableCell>
                                                    <div className="flex items-center gap-2">
                                                        {getStatusIcon(post.post?.status)}
                                                        <Badge className={getStatusColor(post.post?.status)}>{post.post?.status}</Badge>
                                                    </div>
                                                </TableCell>
                                                <TableCell className="max-w-xs">
                                                    <div>
                                                        <p className="font-medium truncate">{post.post?.topic}</p>
                                                        {post.post?.status === PostStatus.FAILED && post.post.reason && (
                                                            <p className="text-xs text-red-600 mt-1">{post.post.reason}</p>
                                                        )}
                                                    </div>
                                                </TableCell>
                                                <TableCell>r/{post.sourceName}</TableCell>
                                                <TableCell className="text-sm">{formatProtoTimestampUTC(post.post?.createdAt)}</TableCell>
                                                <TableCell>
                                                    <div className="text-sm">
                                                        {post.post?.status === PostStatus.SENT && post.post.scheduledAt ? (
                                                            <span>Posted: {formatProtoTimestampUTC(post.post.scheduledAt)}</span>
                                                        ) : post.post?.status === PostStatus.SCHEDULED && post.post.scheduledAt ? (
                                                            <span>Scheduled: {formatProtoTimestampUTC(post.post.scheduledAt)}</span>
                                                        ) : (
                                                            <span className="text-gray-400">-</span>
                                                        )}
                                                    </div>
                                                </TableCell>
                                                {/*<TableCell>*/}
                                                {/*    {post.engagement ? (*/}
                                                {/*        <div className="flex gap-3 text-sm">*/}
                                                {/*            <div className="flex items-center gap-1">*/}
                                                {/*                <Heart className="h-3 w-3" />*/}
                                                {/*                {post.engagement.upvotes}*/}
                                                {/*            </div>*/}
                                                {/*            <div className="flex items-center gap-1">*/}
                                                {/*                <MessageSquare className="h-3 w-3" />*/}
                                                {/*                {post.engagement.comments}*/}
                                                {/*            </div>*/}
                                                {/*        </div>*/}
                                                {/*    ) : (*/}
                                                {/*        <span className="text-gray-400 text-sm">-</span>*/}
                                                {/*    )}*/}
                                                {/*</TableCell>*/}
                                                <TableCell>
                                                    <div className="flex gap-1">
                                                        {[PostStatus.SENT].includes(post.post?.status as PostStatus) && (
                                                            <Button variant="outline" size="sm">
                                                                <Eye className="h-3 w-3" />
                                                            </Button>
                                                        )}
                                                        {post.post?.status === PostStatus.FAILED && (
                                                            <Button variant="outline" size="sm">
                                                                <RefreshCw className="h-3 w-3" />
                                                            </Button>
                                                        )}
                                                        {([PostStatus.SCHEDULED].includes(post.post?.status as PostStatus)) && (
                                                            <Button variant="outline" size="sm" onClick={() => handleDeletePost(post.post)}>
                                                                <Trash2 className="h-4 w-4 text-destructive" />
                                                            </Button>
                                                        )}
                                                        {([PostStatus.CREATED, PostStatus.PROCESSING].includes(post.post?.status as PostStatus)) && (
                                                            <Button variant="outline" size="sm" onClick={() => handleEditPost(post.post)}>
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
                )
            }
        </>
    );
}
