'use client';
import {
    Eye, Edit3,
    AlertCircle, CheckCircle,
    Clock, Loader2, Trash2
} from "lucide-react";
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow
} from "@/components/ui/table";
import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { routes } from "@doota/ui-core/routing";
import { Post, PostDetail } from "@doota/pb/doota/core/v1/post_pb";
import { portalClient } from "@/services/grpc";
import { useAppDispatch } from "@/store/hooks";
import { setPost } from "@/store/PostCreation/PostCreationSlice";
import toast from "react-hot-toast";
import { getConnectError } from "@/utils/error";

export enum PostStatus {
    CREATED = "CREATED",
    SENT = "SENT",
    FAILED = "FAILED",
    SCHEDULED = "SCHEDULED",
}

export default function Posts() {
    const router = useRouter();
    const dispatch = useAppDispatch();
    const [isLoading, setIsLoading] = useState(false)
    const [isPostApiCall, setIsPostApiCall] = useState(false)
    const [posts, setPosts] = useState<PostDetail[]>([])

    useEffect(() => {
        setIsLoading(true)
        Promise.all([
            portalClient.getPosts({}).then(res => setPosts(res.posts)),
        ])
            .catch(err => console.error('Error fetching data:', err))
            .finally(() => setIsLoading(false))
    }, [isPostApiCall])

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
            default:
                return null;
        }
    };

    const getStatusColor = (status?: string) => {
        switch (status) {
            case PostStatus.SENT:
                return "bg-green-100 text-green-800 hover:bg-green-100";
            case PostStatus.SCHEDULED:
                return "bg-blue-100 text-blue-800 hover:bg-blue-100";
            case PostStatus.FAILED:
                return "bg-red-100 text-red-800 hover:bg-red-100";
            case PostStatus.CREATED:
            default:
                return "bg-gray-100 text-gray-800 hover:bg-gray-100";
        }
    };


    const formatProtoTimestampUTC = (
        timestamp?: { seconds: bigint; nanos: number }
    ): string => {
        if (!timestamp) return "-";

        const millis = Number(timestamp.seconds) * 1000 + Math.floor(timestamp.nanos / 1_000_000);
        const date = new Date(millis);

        return date.toLocaleString("en-IN", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            hour12: true,
            // Removed timeZone: "UTC" to use local time
        }).replace(/am|pm/, (match) => match.toUpperCase());;
    };



    const handleEditPost = (post: Post | undefined) => {
        if (post) {
            dispatch(setPost(post));
            router.push(routes.new.postCreationHub.editor);
        }
    };

    const handleDeletePost = async (post: Post | undefined) => {
        try {
            const res = await portalClient.deletePost({ id: post?.id || '' });
            toast.success("Post deleted successfully!");
            setIsPostApiCall(p => !p);
        }
        catch (err: any) {
            toast.error(getConnectError(err));
        }
    }

    return (
        <>
            {
                isLoading ? (
                    <div className="flex justify-center items-center h-screen">
                        <Loader2 className="animate-spin" size={35} />
                    </div>
                )
                    : (
                        <div className="p-6">
                            <div className="mb-6">
                                <div>
                                    <h1 className="text-2xl font-bold">Posts Management</h1>
                                    <p className="text-muted-foreground">Manage your drafts, scheduled posts, and view engagement</p>
                                </div>
                            </div>
                            {posts.length === 0 ? (
                                <Card>
                                    <CardContent className="flex flex-col items-center justify-center py-12 text-center">
                                        <div className="mb-4">
                                            <Edit3 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                                            <h3 className="text-lg font-medium mb-2">No posts created yet</h3>
                                            <p className="text-muted-foreground max-w-md">
                                                Get started by creating your first AI-generated Reddit post. You can use insights or create custom posts.
                                            </p>
                                        </div>
                                        <Button onClick={() => router.push("/post-creation-hub/create")}>
                                            <Edit3 className="h-4 w-4 mr-2" />
                                            Create Your First Post
                                        </Button>
                                    </CardContent>
                                </Card>
                            ) : (
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
                                                                <Badge
                                                                    className={`${getStatusColor(post.post?.status)} hover:shadow-none`}
                                                                >
                                                                    {post.post?.status == PostStatus.CREATED ? "Draft" : post.post?.status}
                                                                </Badge>
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
                                                                {post.post?.status === PostStatus.SENT ? (
                                                                    <span>Posted: {formatProtoTimestampUTC(post.post.scheduledAt)}</span>
                                                                ) : post.post?.status === PostStatus.SCHEDULED ? (
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
                                                                {/*{post.post?.status === PostStatus.FAILED && (*/}
                                                                {/*    <Button variant="outline" size="sm">*/}
                                                                {/*        <RefreshCw className="h-3 w-3" />*/}
                                                                {/*    </Button>*/}
                                                                {/*)}*/}
                                                                {post.post?.status as PostStatus && (
                                                                    <Button variant="outline" size="sm" onClick={() => handleEditPost(post.post)}>
                                                                        <Edit3 className="h-3 w-3" />
                                                                    </Button>
                                                                )}
                                                                {[PostStatus.SENT].includes(post.post?.status as PostStatus) && (
                                                                    <Button
                                                                        variant="outline"
                                                                        size="sm"
                                                                        onClick={() => window.open(post.postUrl, '_blank')}
                                                                    >
                                                                        <Eye className="h-3 w-3" />
                                                                    </Button>
                                                                )}
                                                                {([PostStatus.SCHEDULED, PostStatus.CREATED].includes(post.post?.status as PostStatus)) && (
                                                                    <Button variant="outline" size="sm" onClick={() => handleDeletePost(post.post)}>
                                                                        <Trash2 className="h-4 w-4 text-destructive" />
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
                            )}
                        </div>
                    )
            }
        </>
    );
}
