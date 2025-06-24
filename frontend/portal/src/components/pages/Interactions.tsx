"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import toast from "react-hot-toast";
import {
    MessageSquare,
    MessageCircle,
    Calendar,
    AlertCircle,
    CheckCircle,
    Clock,
    ExternalLink,
    Eye,
    Loader2,
    Trash,
} from "lucide-react";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import {
    LeadInteraction,
    LeadInteractionStatus,
    LeadInteractionType,
} from "@doota/pb/doota/core/v1/core_pb";
import { getFormattedDate } from "@/utils/format";
import { CollapsibleText } from "../Html/HtmlRenderer";

export default function Interaction() {
    const { portalClient } = useClientsContext();
    const [interactions, setInteractions] = useState<LeadInteraction[]>([]);
    const [isFetching, setIsFetching] = useState<boolean>(false);
    const [activeTab, setActiveTab] = useState<LeadInteractionType>(LeadInteractionType.LEAD_INTERACTION_COMMENT);
    const [filter, setFilter] = useState<LeadInteractionStatus>(
        LeadInteractionStatus.UNSPECIFIED
    );

    useEffect(() => {
        setIsFetching(true);
        portalClient
            .getLeadInteractions({})
            .then((res) => {
                setInteractions(res.interactions);
            })
            .catch((err) => {
                console.error("Error fetching integrations:", err);
            })
            .finally(() => {
                setIsFetching(false);
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

    const filteredInteractions = interactions
        .filter((interaction) =>
            activeTab === LeadInteractionType.LEAD_INTERACTION_DM
                ? interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_DM
                : interaction.interactionType ===
                LeadInteractionType.LEAD_INTERACTION_COMMENT
        )
        .filter((interaction) =>
            filter === LeadInteractionStatus.UNSPECIFIED
                ? true
                : interaction.status === filter
        );

    const getStatusIcon = (status: LeadInteractionStatus) => {
        switch (status) {
            case LeadInteractionStatus.SENT:
                return <CheckCircle className="h-4 w-4 text-green-500" />;
            case LeadInteractionStatus.CREATED:
            case LeadInteractionStatus.PROCESSING:
                return <Clock className="h-4 w-4 text-blue-500" />;
            case LeadInteractionStatus.FAILED:
            case LeadInteractionStatus.REMOVED:
                return <AlertCircle className="h-4 w-4 text-red-500" />;
            default:
                return null;
        }
    };

    const getStatusColor = (status: LeadInteractionStatus) => {
        switch (status) {
            case LeadInteractionStatus.SENT:
                return "bg-green-100 text-green-800 hover:bg-green-100";
            case LeadInteractionStatus.CREATED:
            case LeadInteractionStatus.PROCESSING:
                return "bg-blue-100 text-blue-800 hover:bg-blue-100";
            case LeadInteractionStatus.FAILED:
                return "bg-red-100 text-red-800 hover:bg-red-100";
            default:
                return "bg-gray-100 text-gray-800 hover:bg-gray-100";
        }
    };

    const getTypeIcon = (type: LeadInteractionType) =>
        type === LeadInteractionType.LEAD_INTERACTION_DM ? (
            <MessageSquare className="h-4 w-4" />
        ) : (
            <MessageCircle className="h-4 w-4" />
        );

    const handleRemoveInteraction = async (id: string) => {
        try {
            await portalClient.updateLeadInteractionStatus({
                interactionId: id,
                status: LeadInteractionStatus.REMOVED,
            });
            setInteractions((prevInteractions) =>
                prevInteractions.map((interaction) =>
                    interaction.id === id
                        ? {
                            ...interaction,
                            status: LeadInteractionStatus.REMOVED,
                            reason: "Manually removed",
                        }
                        : interaction
                )
            );
        } catch (err: any) {
            console.error("Error updating interaction status to REMOVED:", err);
            const message =
                err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
        }
    };

    const getViewUrl = (interaction: LeadInteraction) => {
        if (interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_COMMENT) {
            return interaction.leadMetadata?.automatedCommentUrl;
        } else if (interaction.interactionType === LeadInteractionType.LEAD_INTERACTION_DM) {
            return interaction.leadMetadata?.dmUrl;
        }
        return "#";
    };

    return (
        <div className="flex-1 space-y-4 p-4 pt-6 md:p-8">
            <div className="flex items-center justify-between space-y-2">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">
                        {activeTab === LeadInteractionType.LEAD_INTERACTION_DM ? "Direct Messages" : "Comments"}
                    </h2>
                    <p className="text-muted-foreground">
                        {activeTab === LeadInteractionType.LEAD_INTERACTION_DM
                            ? "Track and manage your automated DMs"
                            : "Track and manage your automated comments"}
                    </p>
                </div>
            </div>

            <Card>
                <CardContent>
                    {isFetching ? (
                        <div className="flex justify-center items-center my-14">
                            <Loader2 className="animate-spin" size={35} />
                        </div>
                    ) : (
                        <>
                            {/* Tabs */}
                            <div className="flex space-x-2 mt-6 mb-4">
                                <Button
                                    variant={activeTab === LeadInteractionType.LEAD_INTERACTION_COMMENT ? "default" : "outline"}
                                    onClick={() => {
                                        setActiveTab(LeadInteractionType.LEAD_INTERACTION_COMMENT);
                                        setFilter(LeadInteractionStatus.UNSPECIFIED);
                                    }}
                                >
                                    Comments
                                </Button>
                                <Button
                                    variant={activeTab === LeadInteractionType.LEAD_INTERACTION_DM ? "default" : "outline"}
                                    onClick={() => {
                                        setActiveTab(LeadInteractionType.LEAD_INTERACTION_DM);
                                        setFilter(LeadInteractionStatus.UNSPECIFIED);
                                    }}
                                >
                                    DMs
                                </Button>
                            </div>

                            {/* Status filter */}
                            <div className="flex gap-2 mb-6">
                                <Select
                                    value={filter.toString()}
                                    onValueChange={(value) =>
                                        setFilter(Number(value) as LeadInteractionStatus)
                                    }
                                >
                                    <SelectTrigger className="w-[180px]">
                                        <SelectValue placeholder="Filter by status" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem
                                            value={LeadInteractionStatus.UNSPECIFIED.toString()}
                                        >
                                            All ({filteredInteractions.length})
                                        </SelectItem>
                                        <SelectItem value={LeadInteractionStatus.SENT.toString()}>
                                            Sent (
                                            {
                                                filteredInteractions.filter(
                                                    (i) => i.status === LeadInteractionStatus.SENT
                                                ).length
                                            }
                                            )
                                        </SelectItem>
                                        <SelectItem
                                            value={LeadInteractionStatus.CREATED.toString()}
                                        >
                                            Scheduled (
                                            {
                                                filteredInteractions.filter(
                                                    (i) =>
                                                        i.status === LeadInteractionStatus.CREATED ||
                                                        i.status === LeadInteractionStatus.PROCESSING
                                                ).length
                                            }
                                            )
                                        </SelectItem>
                                        <SelectItem
                                            value={LeadInteractionStatus.FAILED.toString()}
                                        >
                                            Failed (
                                            {
                                                filteredInteractions.filter(
                                                    (i) => i.status === LeadInteractionStatus.FAILED
                                                ).length
                                            }
                                            )
                                        </SelectItem>
                                        <SelectItem
                                            value={LeadInteractionStatus.REMOVED.toString()}
                                        >
                                            Removed (
                                            {
                                                filteredInteractions.filter(
                                                    (i) => i.status === LeadInteractionStatus.REMOVED
                                                ).length
                                            }
                                            )
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Interaction list */}
                            <div className="grid gap-4">
                                {filteredInteractions.map((interaction) => (
                                    <Card key={interaction.id} className="p-4">
                                        <div className="flex items-start justify-between gap-4">
                                            <div className="flex-1 space-y-3">
                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center gap-2">
                                                        {getTypeIcon(interaction.interactionType)}
                                                        <span className="font-medium">
                                                            {getInteractionLabel(interaction.interactionType)}
                                                        </span>
                                                    </div>
                                                    <Badge className={getStatusColor(interaction.status)}>
                                                        <div className="flex items-center gap-2">
                                                            {getStatusIcon(interaction.status)}
                                                            {getInteractionStatusLabel(interaction.status)}
                                                        </div>
                                                    </Badge>
                                                </div>

                                                <div className="bg-gray-50 p-3 rounded-md">
                                                    <CollapsibleText
                                                        text={getInteractionMessage(interaction)}
                                                    />
                                                </div>

                                                {(interaction.status === LeadInteractionStatus.FAILED ||
                                                    interaction.status ===
                                                    LeadInteractionStatus.REMOVED) &&
                                                    interaction.reason && (
                                                        <div className="bg-red-50 border border-red-200 p-3 rounded-md">
                                                            <div className="flex items-center gap-2 text-red-800">
                                                                <AlertCircle className="h-4 w-4" />
                                                                <span className="font-medium">
                                                                    Failure Reason:
                                                                </span>
                                                            </div>
                                                            <p className="text-sm text-red-700 mt-1">
                                                                {interaction.reason}
                                                            </p>
                                                        </div>
                                                    )}

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
                                                    {interaction.interactionType ===
                                                        LeadInteractionType.LEAD_INTERACTION_DM && (
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
                                                            <span className="text-gray-500 italic">
                                                                No post URL
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>

                                                <div className="flex items-center gap-1 text-sm text-muted-foreground">
                                                    <Calendar className="h-3 w-3" />
                                                    {getFormattedDate(interaction.scheduledAt)}
                                                </div>
                                            </div>

                                            <div className="flex flex-col gap-2">
                                                {interaction.status === LeadInteractionStatus.SENT && (
                                                    <a
                                                        href={getViewUrl(interaction) ?? "#"}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                    >
                                                        <Button
                                                            variant="outline"
                                                            size="sm"
                                                            className="flex items-center gap-1"
                                                        >
                                                            <Eye className="h-3 w-3" />
                                                            View {getInteractionLabel(interaction.interactionType)}
                                                        </Button>
                                                    </a>
                                                )}

                                                {interaction.status === LeadInteractionStatus.CREATED && (
                                                    <Button
                                                        variant="destructive"
                                                        size="icon"
                                                        className="h-8 w-8"
                                                        onClick={() => handleRemoveInteraction(interaction.id)}
                                                    >
                                                        <Trash className="h-4 w-4" />
                                                    </Button>
                                                )}
                                            </div>
                                        </div>
                                    </Card>
                                ))}
                            </div>

                            {filteredInteractions.length === 0 && !isFetching && (
                                <div className="text-center py-8">
                                    <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                                    <h3 className="text-lg font-medium mb-2">
                                        No interactions found
                                    </h3>
                                    <p className="text-muted-foreground">
                                        No interactions found.
                                    </p>
                                </div>
                            )}
                        </>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
