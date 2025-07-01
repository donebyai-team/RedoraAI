"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TrendingUp, Search, MessageSquare, BarChart3, PenTool, ExternalLink, Loader2 } from "lucide-react";
import { PostInsight } from "@doota/pb/doota/core/v1/insight_pb";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { getFormattedDate } from "@/utils/format";


export default function Insights() {
    const { portalClient } = useClientsContext();
    const [insights, setInsights] = useState<PostInsight[]>([]);
    const [searchTerm, setSearchTerm] = useState("");
    const [sentimentFilter, setSentimentFilter] = useState("all");
    const [sourceFilter, setSourceFilter] = useState("all");
    const [isFetching, setIsFetching] = useState<boolean>(false);


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
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const handleCreatePost = (insight: PostInsight) => {
        console.log("Create post for insight:", insight.topic);
        // TODO: Implement post creation functionality
    };

    const getRedditCommentUrl = (postId: string, commentId: string) => {
        return `https://reddit.com/comments/${postId}/_/${commentId}`;
    };

    const getSentimentColor = (sentiment: string) => {
        switch (sentiment) {
            case "positive":
                return "bg-green-100 text-green-800";
            case "negative":
                return "bg-red-100 text-red-800";
            case "neutral":
                return "bg-gray-100 text-gray-800";
            default:
                return "bg-gray-100 text-gray-800";
        }
    };

    const getRelevancyColor = (score: number) => {
        if (score >= 90) return "text-red-600";
        if (score >= 70) return "text-orange-600";
        return "text-yellow-600";
    };

    const filteredInsights = insights.filter(insight => {
        const matchesSearch = insight.topic.toLowerCase().includes(searchTerm.toLowerCase()) ||
            insight.postTitle.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesSentiment = sentimentFilter === "all" || insight.sentiment === sentimentFilter;
        const matchesSource = sourceFilter === "all" || insight.source.toLowerCase() === sourceFilter.toLowerCase();

        return matchesSearch && matchesSentiment && matchesSource;
    });

    return (
        <div>
            {isFetching ? (
                <div className="flex justify-center items-center my-14">
                    <Loader2 className="animate-spin" size={35} />
                </div>
            ) : (
                <>
                    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
                        <div className="flex items-center justify-between space-y-2">
                            <div>
                                <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                                    <TrendingUp className="h-8 w-8" />
                                    Weekly Insights
                                </h2>
                                <p className="text-muted-foreground mt-2">
                                    Discover trending topics and discussions across communities
                                </p>
                            </div>
                            <div className="flex items-center space-x-2">
                                <Badge variant="secondary" className="px-3 py-1">
                                    {filteredInsights.length} insights
                                </Badge>
                            </div>
                        </div>

                        {/* Filters */}
                        <Card>
                            <CardContent className="pt-6">
                                <div className="flex flex-col md:flex-row gap-4">
                                    <div className="flex-1">
                                        <div className="relative">
                                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                                            <Input
                                                placeholder="Search insights by topic or post title..."
                                                value={searchTerm}
                                                onChange={(e) => setSearchTerm(e.target.value)}
                                                className="pl-8"
                                            />
                                        </div>
                                    </div>
                                    <Select value={sentimentFilter} onValueChange={setSentimentFilter}>
                                        <SelectTrigger className="w-[180px]">
                                            <SelectValue placeholder="Filter by sentiment" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="all">All Sentiments</SelectItem>
                                            <SelectItem value="positive">Positive</SelectItem>
                                            <SelectItem value="negative">Negative</SelectItem>
                                            <SelectItem value="neutral">Neutral</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <Select value={sourceFilter} onValueChange={setSourceFilter}>
                                        <SelectTrigger className="w-[160px]">
                                            <SelectValue placeholder="Filter by source" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="all">All Sources</SelectItem>
                                            <SelectItem value="subreddit">Reddit</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Insights Grid */}
                        <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
                            {filteredInsights.map((insight) => (
                                <Card key={insight.id} className="hover:shadow-md transition-shadow">
                                    <CardHeader className="pb-3">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <CardTitle className="text-lg leading-tight mb-2">
                                                    {insight.topic}
                                                </CardTitle>
                                                <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
                                                    <MessageSquare className="h-4 w-4" />
                                                    <span className="truncate max-w-md">{insight.postTitle}</span>
                                                </div>
                                            </div>
                                            <div className="flex flex-col items-end gap-2">
                                                <div className="flex items-center gap-1">
                                                    <BarChart3 className={`h-4 w-4 ${getRelevancyColor(insight.relevancyScore)}`} />
                                                    <span className={`font-semibold ${getRelevancyColor(insight.relevancyScore)}`}>
                                                        {insight.relevancyScore}%
                                                    </span>
                                                </div>
                                                <Badge className={getSentimentColor(insight.sentiment)}>
                                                    {insight.sentiment}
                                                </Badge>
                                            </div>
                                        </div>
                                    </CardHeader>
                                    <CardContent className="pt-0">
                                        <div className="space-y-4">
                                            <div>
                                                <h4 className="font-medium text-sm mb-2">Key Highlights:</h4>
                                                <div className="text-sm text-muted-foreground bg-muted/50 p-3 rounded-md">
                                                    {insight.highlights.split('\n').map((highlight, index) => (
                                                        <p key={index} className="mb-1 last:mb-0">
                                                            {highlight}
                                                        </p>
                                                    ))}
                                                </div>
                                            </div>

                                            <div>
                                                <h4 className="font-medium text-sm mb-2">Analysis:</h4>
                                                <p className="text-sm text-muted-foreground leading-relaxed">
                                                    {insight.cot}
                                                </p>
                                            </div>

                                            {insight.highlightedComments.length > 0 && (
                                                <div>
                                                    <h4 className="font-medium text-sm mb-2">References:</h4>
                                                    <div className="flex flex-wrap gap-2">
                                                        {insight.highlightedComments.map((commentId, index) => (
                                                            <a
                                                                key={commentId}
                                                                href={getRedditCommentUrl(insight.postId, commentId)}
                                                                target="_blank"
                                                                rel="noopener noreferrer"
                                                                className="inline-flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 hover:underline bg-blue-50 px-2 py-1 rounded-md transition-colors"
                                                            >
                                                                Comment {index + 1}
                                                                <ExternalLink className="h-3 w-3" />
                                                            </a>
                                                        ))}
                                                    </div>
                                                </div>
                                            )}

                                            <div className="flex items-center justify-between pt-2 border-t">
                                                <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-muted-foreground">
                                                    <span>Source: {insight.source}</span>
                                                    <span>Post: {getFormattedDate(insight.postCreatedAt)}</span>
                                                    <span>Analyzed: {getFormattedDate(insight.createdAt!)}</span>
                                                </div>
                                                <Button
                                                    size="sm"
                                                    onClick={() => handleCreatePost(insight)}
                                                    className="h-8 px-3"
                                                >
                                                    <PenTool className="h-3 w-3 mr-1" />
                                                    Create Post
                                                </Button>
                                            </div>
                                        </div>
                                    </CardContent>
                                </Card>
                            ))}
                        </div>

                        {filteredInsights.length === 0 && (
                            <Card>
                                <CardContent className="flex items-center justify-center py-12">
                                    <div className="text-center">
                                        <TrendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                                        <h3 className="text-lg font-medium mb-2">No insights found</h3>
                                        <p className="text-muted-foreground">
                                            Try adjusting your search terms or filters to see more results.
                                        </p>
                                    </div>
                                </CardContent>
                            </Card>
                        )}
                    </div>
                </>
            )}
        </div>
    );
}
