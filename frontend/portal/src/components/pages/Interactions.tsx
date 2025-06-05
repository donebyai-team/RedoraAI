"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Collapsible, CollapsibleTrigger } from "@/components/ui/collapsible";
import { MessageSquare, MessageCircle, Calendar, AlertCircle, CheckCircle, Clock, ExternalLink, Eye, X, ChevronDown, ChevronUp } from "lucide-react";
import { setLoading, setError } from "@/store/Source/sourceSlice";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { LeadInteraction, LeadInteractionStatus, LeadInteractionType } from "@doota/pb/doota/core/v1/core_pb";
import { getFormattedDate } from "@/utils/format";
import { RootState } from "@/store/store";

const CollapsibleText = ({ text }: { text: string }) => {
    const [isOpen, setIsOpen] = useState(false);
    const shouldCollapse = text.length > 100; // Collapse if text is longer than 100 characters

    if (!shouldCollapse) {
        return (
            <p className="text-sm text-gray-700">
                {text}
            </p>
        );
    }

    return (
        <Collapsible open={isOpen} onOpenChange={setIsOpen}>
            <div className="space-y-2">
                <p className="text-sm text-gray-700">
                    {isOpen ? text : `${text.substring(0, 100)}...`}
                </p>
                <CollapsibleTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-auto p-0 text-blue-600 hover:text-blue-800">
                        <span className="flex items-center gap-1 text-xs">
                            {isOpen ? (
                                <>
                                    Read less <ChevronUp className="h-3 w-3" />
                                </>
                            ) : (
                                <>
                                    Read more <ChevronDown className="h-3 w-3" />
                                </>
                            )}
                        </span>
                    </Button>
                </CollapsibleTrigger>
            </div>
        </Collapsible>
    );
};

export default function Interaction() {
    const { portalClient } = useClientsContext();
    const dispatch = useAppDispatch();
    const [interactions, setInteractions] = useState<LeadInteraction[]>([]);
    const { isLoading } = useAppSelector((state: RootState) => state.lead);

    useEffect(() => {
        dispatch(setLoading(true));
        portalClient.getLeadInteractions({})
            .then((res) => {
                setInteractions(res.interactions);
            })
            .catch((err) => {
                console.error("Error fetching integrations:", err);
                dispatch(setError('Failed to fetch interactions'));
            })
            .finally(() => {
                dispatch(setLoading(false));
            });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const getInteractionLabel = (type: LeadInteractionType) => {
        switch (type) {
            case LeadInteractionType.LEAD_INTERACTION_COMMENT:
                return "Comment";
            case LeadInteractionType.LEAD_INTERACTION_DM:
                return "DM";
            default:
                return "";
        }
    };

    const getInteractionStatusLabel = (type: LeadInteractionStatus) => {
        switch (type) {
            case LeadInteractionStatus.SENT:
                return "Sent";
            case LeadInteractionStatus.REMOVED:
                return "Removed";
            case LeadInteractionStatus.FAILED:
                return "Failed";
            case LeadInteractionStatus.CREATED:
                return "Scheduled";
            case LeadInteractionStatus.PROCESSING:
                return "Scheduled";
            default:
                return "";
        }
    };

    const getInteractionMessage = (interaction: LeadInteraction): string => {
        switch (interaction.interactionType) {
            case LeadInteractionType.LEAD_INTERACTION_COMMENT:
                return interaction.leadMetadata?.suggestedComment ?? "";
            case LeadInteractionType.LEAD_INTERACTION_DM:
                return interaction.leadMetadata?.suggestedDm ?? "";
            default:
                return "";
        }
    };


    const [filter, setFilter] = useState<LeadInteractionStatus>(LeadInteractionStatus.UNSPECIFIED);

    const filteredInteractions = filter === LeadInteractionStatus.UNSPECIFIED
        ? interactions
        : interactions.filter(interaction => interaction.status === filter);

    const getStatusIcon = (status: LeadInteractionStatus) => {
        switch (status) {
            case LeadInteractionStatus.SENT:
                return <CheckCircle className="h-4 w-4 text-green-500" />;
            case LeadInteractionStatus.CREATED:
                return <Clock className="h-4 w-4 text-blue-500" />;
            case LeadInteractionStatus.PROCESSING:
                return <Clock className="h-4 w-4 text-blue-500" />;
            case LeadInteractionStatus.FAILED:
                return <AlertCircle className="h-4 w-4 text-red-500" />;
            case LeadInteractionStatus.REMOVED:
                return <AlertCircle className="h-4 w-4 text-red-500" />;
            default:
                return null;
        }
    };

    const getStatusColor = (status: LeadInteractionStatus) => {
        switch (status) {
            case LeadInteractionStatus.SENT:
                return 'bg-green-100 text-green-800 hover:bg-green-100';
            case LeadInteractionStatus.CREATED:
                return 'bg-blue-100 text-blue-800 hover:bg-blue-100';
            case LeadInteractionStatus.PROCESSING:
                return 'bg-blue-100 text-blue-800 hover:bg-blue-100';
            case LeadInteractionStatus.FAILED:
                return 'bg-red-100 text-red-800 hover:bg-red-100';
            default:
                return 'bg-gray-100 text-gray-800 hover:bg-gray-100';
        }
    };

    const getTypeIcon = (type: LeadInteractionType) => {
        return type === LeadInteractionType.LEAD_INTERACTION_DM ? <MessageSquare className="h-4 w-4" /> : <MessageCircle className="h-4 w-4" />;
    };

    const handleCancelScheduled = (id: string) => {
        console.log('Canceling scheduled interaction:', id);
        // TODO: Implement cancel logic
    };

    const getViewUrl = (interaction: LeadInteraction) => {
        if (interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_COMMENT) {
            return interaction.leadMetadata?.automatedCommentUrl
        } else if (interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_DM) {
            return interaction.leadMetadata?.dmUrl
        }
        return "#";
    };

    return (
        <div className="flex-1 space-y-4 p-4 pt-6 md:p-8">
            <div className="flex items-center justify-between space-y-2">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Interactions</h2>
                    <p className="text-muted-foreground">
                        Track and manage your automated DMs and comments
                    </p>
                </div>
            </div>

            <Card>
                {/* <CardHeader>
                    <CardTitle>Automation Actions</CardTitle>
                    <CardDescription>
                        All your automated interactions with Reddit users
                    </CardDescription>
                </CardHeader> */}
                <CardContent>
                    {/* Filter dropdown */}
                    <div className="flex gap-2 mb-6 mt-6">
                        <Select
                            value={filter.toString()}
                            onValueChange={(value) => setFilter(Number(value) as LeadInteractionStatus)}
                        >
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Filter by status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value={LeadInteractionStatus.UNSPECIFIED.toString()}>
                                    All ({interactions.length})
                                </SelectItem>
                                <SelectItem value={LeadInteractionStatus.SENT.toString()}>
                                    Sent ({interactions.filter(i => i.status === LeadInteractionStatus.SENT).length})
                                </SelectItem>
                                <SelectItem value={LeadInteractionStatus.CREATED.toString()}>
                                    Scheduled ({interactions.filter(i =>
                                        i.status === LeadInteractionStatus.CREATED || i.status === LeadInteractionStatus.PROCESSING
                                    ).length})
                                </SelectItem>
                                <SelectItem value={LeadInteractionStatus.FAILED.toString()}>
                                    Failed ({interactions.filter(i => i.status === LeadInteractionStatus.FAILED).length})
                                </SelectItem>
                                <SelectItem value={LeadInteractionStatus.REMOVED.toString()}>
                                    Removed ({interactions.filter(i => i.status === LeadInteractionStatus.REMOVED).length})
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    {/* Interactions grid */}
                    <div className="grid gap-4">
                        {filteredInteractions.map((interaction) => (
                            <Card key={interaction.id} className="p-4">
                                <div className="flex items-start justify-between gap-4">
                                    <div className="flex-1 space-y-3">
                                        {/* Header with type and status */}
                                        <div className="flex items-center justify-between">
                                            <div className="flex items-center gap-2">
                                                {getTypeIcon(interaction.interactionType)}
                                                <span className="font-medium">{getInteractionLabel(interaction.interactionType)}</span>
                                            </div>
                                            <Badge className={getStatusColor(interaction.status)}>
                                                <div className="flex items-center gap-2">
                                                    {getStatusIcon(interaction.status)}
                                                    {getInteractionStatusLabel(interaction.status)}
                                                </div>
                                            </Badge>
                                        </div>

                                        {/* Message preview with collapsible text */}
                                        <div className="bg-gray-50 p-3 rounded-md">
                                            <CollapsibleText text={getInteractionMessage(interaction)} />
                                        </div>

                                        {/* Failure reason for failed interactions */}
                                        {(interaction.status === LeadInteractionStatus.FAILED || interaction.status === LeadInteractionStatus.REMOVED) && interaction.reason && (
                                            <div className="bg-red-50 border border-red-200 p-3 rounded-md">
                                                <div className="flex items-center gap-2 text-red-800">
                                                    <AlertCircle className="h-4 w-4" />
                                                    <span className="font-medium">Failure Reason:</span>
                                                </div>
                                                <p className="text-sm text-red-700 mt-1">
                                                    {interaction.reason}
                                                </p>
                                            </div>
                                        )}

                                        {/* User info and post */}
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm text-muted-foreground">
                                            <div>
                                                <span className="font-medium">From:</span>{" "}
                                                <a
                                                    href={`https://www.reddit.com/user/${interaction.from}`}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    className="inline-flex items-center gap-1"
                                                >
                                                    u/{interaction.from}
                                                </a>
                                            </div>
                                            {interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_DM && (
                                                <div>
                                                    <span className="font-medium">To:</span>{" "}
                                                    <a
                                                        href={`https://www.reddit.com/user/${interaction.to}`}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="inline-flex items-center gap-1"
                                                    >
                                                        u/{interaction.to}
                                                    </a>
                                                </div>
                                            )}

                                            <div className="md:col-span-2">
                                                <span className="font-medium">Post:</span>{" "}
                                                {interaction.leadMetadata?.postUrl ? (
                                                    <a
                                                        href={interaction.leadMetadata?.postUrl}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="inline-flex items-center gap-1"
                                                    >
                                                        {interaction.postTitle}
                                                        <ExternalLink className="h-3 w-3" />
                                                    </a>
                                                ) : (
                                                    <span className="text-gray-500 italic">No post URL</span>
                                                )}
                                            </div>
                                        </div>

                                        {/* Timestamp */}
                                        <div className="flex items-center gap-1 text-sm text-muted-foreground">
                                            <Calendar className="h-3 w-3" />
                                            {getFormattedDate(interaction.scheduledAt)}
                                        </div>
                                    </div>

                                    {/* Action buttons */}
                                    <div className="flex flex-col gap-2">
                                        {interaction.status === LeadInteractionStatus.SENT && (
                                            <Dialog>
                                                <a
                                                    href={getViewUrl(interaction) ?? "#"}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                >
                                                    <Button variant="outline" size="sm" className="flex items-center gap-1">
                                                        <Eye className="h-3 w-3" />
                                                        View {getInteractionLabel(interaction.interactionType)}
                                                    </Button>
                                                </a>
                                                {/* <DialogContent className="max-w-2xl">
                                                    <DialogHeader>
                                                        <DialogTitle>
                                                            {interaction.interactionType} to {interaction.to}
                                                        </DialogTitle>
                                                    </DialogHeader>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <label className="text-sm font-medium text-muted-foreground">Message:</label>
                                                            <div className="mt-1 p-3 bg-gray-50 rounded-md">
                                                                <p className="text-sm">{interaction.leadMetadata?.suggestedComment}</p>
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <label className="text-sm font-medium text-muted-foreground">Related Post:</label>
                                                            <div className="mt-1">
                                                                {interaction.leadMetadata?.postUrl ? (
                                                                    <a
                                                                        href={interaction.leadMetadata?.postUrl}
                                                                        target="_blank"
                                                                        rel="noopener noreferrer"
                                                                        className="inline-flex items-center gap-1"
                                                                    >
                                                                        {interaction.postTitle}
                                                                        <ExternalLink className="h-3 w-3" />
                                                                    </a>
                                                                ) : (
                                                                    <span className="text-gray-500 italic">No post URL</span>
                                                                )}
                                                            </div>
                                                        </div>
                                                        <div className="text-xs text-muted-foreground">
                                                            Sent at {getFormattedDate(interaction.scheduledAt)}
                                                        </div>
                                                    </div>
                                                </DialogContent> */}
                                            </Dialog>
                                        )}

                                        {/* {interaction.status === LeadInteractionStatus.CREATED && (
                                            <Button
                                                variant="destructive"
                                                size="sm"
                                                className="flex items-center gap-1"
                                                onClick={() => handleCancelScheduled(interaction.id)}
                                            >
                                                <X className="h-3 w-3" />
                                                Cancel
                                            </Button>
                                        )} */}
                                    </div>
                                </div>
                            </Card>
                        ))}
                    </div>

                    {filteredInteractions.length === 0 && !isLoading && (
                        <div className="text-center py-8">
                            <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-lg font-medium mb-2">No interactions found</h3>
                            <p className="text-muted-foreground">
                                No interactions found.
                            </p>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
