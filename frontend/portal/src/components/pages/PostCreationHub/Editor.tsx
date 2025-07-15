'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import {
    Card, CardContent, CardHeader, CardTitle
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
    ArrowLeft, Calendar, History, Save, Undo, ChevronUp, ChevronDown, Loader2
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
    Collapsible, CollapsibleContent, CollapsibleTrigger
} from '@/components/ui/collapsible'
import {
    Select, SelectContent, SelectItem, SelectTrigger, SelectValue
} from '@/components/ui/select'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { PostRegenerationHistory, Post } from '@doota/pb/doota/core/v1/post_pb'
import { useCreatePost } from '@/components/hooks/useCreatePost'
import { portalClient } from '@/services/grpc'
import toast from 'react-hot-toast'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import { PostInsight } from '@doota/pb/doota/core/v1/insight_pb'
import { Source } from '@doota/pb/doota/core/v1/core_pb'
import {routes} from "@doota/ui-core/routing";
import {PostStatus} from "@/components/pages/PostCreationHub/Posts";
import {setPost} from "@/store/PostCreation/PostCreationSlice";
import {getConnectError} from "@/utils/error";

const editableStatus = [PostStatus.CREATED, PostStatus.SCHEDULED]

const getMinDateTimeLocal = () => {
    const now = new Date()
    return now.toISOString().slice(0, 16)
}

export function formatTimestampToLocalInput(timestamp?: Timestamp): string {
    if (!timestamp || typeof timestamp.seconds === 'undefined') return '';

    const milliseconds =
        Number(timestamp.seconds) * 1000 + Math.floor(timestamp.nanos / 1_000_000);

    const utcDate = new Date(milliseconds);

    const localDate = new Date(utcDate.getTime() - utcDate.getTimezoneOffset() * 60000);

    return localDate.toISOString().slice(0, 16); // "yyyy-MM-ddTHH:mm"
}

export default function PostEditor() {
    const router = useRouter()
    const dispatch = useAppDispatch();
    const { createPost } = useCreatePost()
    const { post } = useAppSelector(state => state.postCreation)


    const [title, setTitle] = useState(post?.topic || '')
    const [content, setContent] = useState(post?.description || '')
    const [scheduledDate, setScheduledDate] = useState('')
    const [isLoading, setIsLoading] = useState(false);
    const [isPostApiCall, setIsPostApiCall] = useState(false)
    const [generationHistory, setGenerationHistory] = useState<PostRegenerationHistory[]>([])
    const [showHistory, setShowHistory] = useState(false)
    const [showEditContext, setShowEditContext] = useState(false)

    const [sources, setSources] = useState<Source[]>([])
    const [insights, setInsights] = useState<PostInsight[]>([])

    const [selectedInsight, setSelectedInsight] = useState(post?.metadata?.settings?.referenceId || '')
    const [customTopic, setCustomTopic] = useState(post?.metadata?.settings?.topic || '')
    const [postDetails, setPostDetails] = useState(post?.metadata?.settings?.context || '')
    const [selectedGoal, setSelectedGoal] = useState(post?.metadata?.settings?.goal || '')
    const [selectedSubreddit, setSelectedSubreddit] = useState(post?.source || '')
    const [selectedTone, setSelectedTone] = useState(post?.metadata?.settings?.tone || '')
    const [isEditable, setIsEditable] = useState(editableStatus.includes(post?.status as PostStatus || ''))
    const [selectedVersionIndex, setSelectedVersionIndex] = useState<number>(0)


    useEffect(() => {
        if(!post?.id) router.back();

       setIsLoading(true);

        Promise.all([
            portalClient.getInsights({}).then(res => setInsights(res.insights)),
            portalClient.getSources({}).then(res => setSources(res.sources)),
        ])
            .catch((err) => console.error('Error fetching data:', err))
            .finally(() => setIsLoading(false));

        return () => {
            dispatch(setPost(null));
        }
    }, []);

    useEffect(() => {
        if (post?.topic && post?.description) {
            setGenerationHistory(post?.metadata?.history ?? [])
        }

        if (post?.scheduledAt) {
            const inputFormatted = formatTimestampToLocalInput(post.scheduledAt);
            setScheduledDate(inputFormatted);
        }

    }, [post])

    const handleInsightSelect = (insightId: string) => {
        setSelectedInsight(insightId)
        const insight = insights.find(i => i.id === insightId)
        if (insight) {
            setCustomTopic(insight.topic)
            setPostDetails(insight.highlights)
        }
    }

    const handleRegenerate = async () => {
        if (!post?.id) return

        setIsPostApiCall(true)
        const res: Post | undefined = await createPost({
            id: post.id,
            sourceId: selectedSubreddit,
            topic: customTopic,
            context: postDetails,
            goal: selectedGoal,
            tone: selectedTone,
            referenceId: selectedInsight ?? null,
        },false,
        )

        if (res) {
            setTitle(res.topic || '')
            setContent(res.description || '')
            setGenerationHistory(res.metadata?.history || [])
        }

        setIsPostApiCall(false)
    }

    const handleSelectFromHistory = (item: PostRegenerationHistory) => {
        setTitle(item.title || '')
        setContent(item.description || '')
        setSelectedInsight(item.postSettings?.referenceId || '')
        setCustomTopic(item.postSettings?.topic || '')
        setPostDetails(item.postSettings?.context || '')
        setSelectedSubreddit(post?.source || '')
        setSelectedGoal(item.postSettings?.goal || '')
        setSelectedTone(item.postSettings?.tone || '')
        setShowHistory(false)
    }

    const handleSchedule = async () => {
        if (!scheduledDate) {
            toast.error("Please select a date and time")
            return
        }

        try {
            const date = new Date(scheduledDate)
            const timestamp: Omit<Timestamp, '$typeName'> = {
                seconds: BigInt(Math.floor(date.getTime() / 1000)),
                nanos: (date.getTime() % 1000) * 1_000_000
            }

            setIsPostApiCall(true)
            await portalClient.schedulePost({
                id: post?.id || '',
                scheduleAt: timestamp,
                version: "v"+(selectedVersionIndex + 1),
            })
            toast.success('Post scheduled successfully!')
            router.push(routes.new.postCreationHub.posts)
        } catch (err: any) {
            toast.error(getConnectError(err));
        } finally {
            setTimeout(() => setIsPostApiCall(false),1000)
        }
    }

    const handleSaveDraft = () => {
        toast.success('Draft saved (stub)')
        router.push(routes.new.postCreationHub.posts)
    }

    const goalOptions = [
        { value: 'karma', label: 'Build Karma' },
        { value: 'feedback', label: 'Get Feedback' },
        { value: 'leads', label: 'Generate Leads' }
    ]

    const toneOptions = [
        { value: 'professional', label: 'Professional' },
        { value: 'casual', label: 'Casual' },
        { value: 'friendly', label: 'Friendly' }
    ]

    const formatTime = (index: number) => `v${generationHistory.length - index}`

    useEffect(() => {
        if (post?.topic && generationHistory.length > 0) {
            const matchIndex = generationHistory.findIndex(h => h.title === post.topic)
            if (matchIndex !== -1) {
                setSelectedVersionIndex(matchIndex)
            }
        }
    }, [generationHistory, post])

    return (
        <div>
            {
                isLoading ? (
                        <div className="flex justify-center items-center h-screen">
                            <Loader2 className="animate-spin" size={35} />
                        </div>
                    )
                    :(
                    <div className="p-6 max-w-6xl mx-auto">
                        {/*<div className="mb-6 flex items-center gap-4">*/}
                        <Button variant="outline" onClick={() => router.back()}>
                            <ArrowLeft className="h-4 w-4 mr-2" /> Back
                        </Button>
                        <div className="my-4">
                            <h1 className="text-2xl font-bold">Edit Generated Post</h1>
                            <p className="text-muted-foreground">Review, edit, and schedule your AI-generated post</p>
                        </div>
                        {/*</div>*/}

                        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                            {/* LEFT: Editor */}
                            <div className="lg:col-span-3">
                                <Card>
                                    <CardHeader className="flex flex-row justify-between items-center">
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
                                    </CardHeader>
                                    <CardContent className="space-y-4">
                                        {showHistory && generationHistory.length > 0 && (
                                            <div className="border p-4 bg-gray-50 rounded-lg">
                                                <h4 className="font-medium mb-3">Generation History</h4>
                                                <div className="space-y-2 max-h-60 overflow-y-auto">
                                                    {[...generationHistory].reverse().map((item, index) => {
                                                        const actualIndex = generationHistory.length - 1 - index; // to map back to real index
                                                        const isSelected = selectedVersionIndex == actualIndex;

                                                        const highlightClass = isSelected
                                                            ? 'bg-gray-200 hover:bg-gray-200 border-gray-400'
                                                            : 'bg-white hover:bg-gray-50';

                                                        return (
                                                            <div
                                                                key={index}
                                                                className={`flex items-center justify-between p-3 rounded border cursor-pointer ${highlightClass}`}
                                                                onClick={() => {
                                                                    handleSelectFromHistory(item);
                                                                    setSelectedVersionIndex(actualIndex);
                                                                }}
                                                            >
                                                                <div className="min-w-0">
                                                                    <p className="font-medium text-sm truncate">{item.title}</p>
                                                                    <p className="text-xs text-muted-foreground truncate">{item.description}</p>
                                                                </div>
                                                                <Badge variant="secondary" className="text-xs">
                                                                    {formatTime(index)}
                                                                </Badge>
                                                            </div>
                                                        );
                                                    })}

                                                </div>
                                            </div>
                                        )}

                                        <div>
                                            <Label className='mb-2.5 block'>Title</Label>
                                            <Input disabled={!isEditable} value={title} onChange={e => setTitle(e.target.value)} />
                                        </div>

                                        <div>
                                            <Label className='mb-2.5 block'>Content</Label>
                                            <Textarea
                                                disabled={!isEditable}
                                                className="min-h-[400px]"
                                                value={content}
                                                onChange={e => setContent(e.target.value)}
                                            />
                                        </div>

                                        {/* Collapsible Context Section */}
                                        <Collapsible open={showEditContext} onOpenChange={setShowEditContext}>
                                            <CollapsibleTrigger asChild>
                                                <div className="flex items-center justify-between cursor-pointer p-3 border rounded-lg hover:bg-gray-50">
                                                    <h4 className="font-medium">Edit Post Context</h4>
                                                    <Button variant="ghost" size="sm">
                                                        {showEditContext ? (
                                                            <ChevronUp className="h-4 w-4" />
                                                        ) : (
                                                            <ChevronDown className="h-4 w-4" />
                                                        )}
                                                    </Button>
                                                </div>
                                            </CollapsibleTrigger>
                                            <CollapsibleContent>
                                                <div className="mt-4 space-y-4 bg-gray-50 border rounded-lg p-4">
                                                    {/*Suggested topic*/}
                                                    <div>
                                                        <Label htmlFor="insight-select" className='mb-2.5 block'>Suggested Topics from Insights (Optional)</Label>
                                                        <Select onValueChange={handleInsightSelect} disabled={!isEditable}>
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
                                                                            <span className="truncate max-w-[300px]">{insight.topic}</span>
                                                                        </div>
                                                                    </SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>

                                                    {/* Topic Input */}
                                                    <div>
                                                        <Label htmlFor="topic" className='mb-2.5 block'>Topic</Label>
                                                        <Textarea
                                                            id="topic"
                                                            value={customTopic}
                                                            onChange={(e) => setCustomTopic(e.target.value)}
                                                            placeholder="Enter your topic..."
                                                            className="min-h-[100px] text-base"
                                                            disabled={!isEditable}
                                                        />
                                                    </div>

                                                    {/* Post Details */}
                                                    <div>
                                                        <Label htmlFor="details" className='mb-2.5 block'>Post Details & Context</Label>
                                                        <Textarea
                                                            id="details"
                                                            value={postDetails}
                                                            onChange={(e) => setPostDetails(e.target.value)}
                                                            placeholder="Add specific details, context, examples, or requirements for your post..."
                                                            className="min-h-[250px] text-base"
                                                            disabled={!isEditable}
                                                        />
                                                    </div>
                                                </div>
                                            </CollapsibleContent>
                                        </Collapsible>

                                        <Button variant="outline" onClick={handleRegenerate}
                                                disabled={
                                                    isPostApiCall ||
                                                    !isEditable ||
                                                    !customTopic ||
                                                    !selectedSubreddit ||
                                                    !selectedGoal ||
                                                    !selectedTone ||
                                                    !postDetails
                                                }
                                        >
                                            <Undo className="h-4 w-4 mr-2" />
                                            {isPostApiCall ? 'Regenerating...' : 'Regenerate' }
                                        </Button>
                                    </CardContent>
                                </Card>
                            </div>

                            {/* RIGHT: Settings + Schedule */}
                            <div className="lg:col-span-1 space-y-4">
                                <Card>
                                    <CardHeader>
                                        <CardTitle>Post Settings</CardTitle>
                                    </CardHeader>
                                    <CardContent className="space-y-4">
                                        <div>
                                            <Label className='mb-2.5 block'>Subreddit</Label>
                                            <Select value={selectedSubreddit} onValueChange={setSelectedSubreddit}
                                                    disabled={!isEditable}
                                            >
                                                <SelectTrigger><SelectValue placeholder="Select subreddit" /></SelectTrigger>
                                                <SelectContent>
                                                    {sources.map(src => (
                                                        <SelectItem key={src.id} value={src.id}>{src.name}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>

                                        <div>
                                            <Label className='mb-2.5 block'>Goal</Label>
                                            <Select value={selectedGoal} onValueChange={setSelectedGoal}
                                                    disabled={!isEditable}
                                            >
                                                <SelectTrigger><SelectValue placeholder="Select goal" /></SelectTrigger>
                                                <SelectContent>
                                                    {goalOptions.map(goal => (
                                                        <SelectItem key={goal.value} value={goal.value}>{goal.label}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>

                                        <div>
                                            <Label className='mb-2.5 block'>Tone</Label>
                                            <Select value={selectedTone} onValueChange={setSelectedTone}
                                                    disabled={!isEditable}
                                            >
                                                <SelectTrigger><SelectValue placeholder="Select tone" /></SelectTrigger>
                                                <SelectContent>
                                                    {toneOptions.map(tone => (
                                                        <SelectItem key={tone.value} value={tone.value}>{tone.label}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </CardContent>
                                </Card>

                                <Card>
                                    <CardHeader className="flex flex-row justify-between items-center">
                                        <CardTitle>Post Actions</CardTitle>
                                    </CardHeader>
                                    <CardContent className="space-y-4">
                                        <Label>Schedule Date & Time</Label>
                                        <Input
                                            type="datetime-local"
                                            min={getMinDateTimeLocal()}
                                            value={scheduledDate}
                                            onChange={(e) => setScheduledDate(e.target.value)}
                                            className="w-full pr-10 text-sm appearance-none relative
                                            [&::-webkit-calendar-picker-indicator]:absolute
                                            [&::-webkit-calendar-picker-indicator]:right-2
                                            [&::-webkit-calendar-picker-indicator]:top-1/2
                                            [&::-webkit-calendar-picker-indicator]:-translate-y-1/2
                                            [&::-webkit-calendar-picker-indicator]:cursor-pointer
                                            [&::-webkit-calendar-picker-indicator]:h-5
                                            [&::-webkit-calendar-picker-indicator]:w-5
                                            [&::-webkit-calendar-picker-indicator]:bg-transparent
                                            [&::-webkit-calendar-picker-indicator]:opacity-100"
                                        />
                                        {/*<Button onClick={handleSaveDraft} variant="outline" className="w-full">*/}
                                        {/*    <Save className="h-4 w-4 mr-2" />*/}
                                        {/*    Save as Draft*/}
                                        {/*</Button>*/}

                                        <Button onClick={handleSchedule} className="w-full"
                                            disabled={isPostApiCall || !isEditable}
                                        >
                                            <Calendar className="h-4 w-4 mr-2" />
                                            {post?.scheduledAt ? "Update Post" :  "Schedule Post"}
                                        </Button>
                                    </CardContent>
                                </Card>
                            </div>
                        </div>
                    </div>
                )
            }
        </div>
    )
}