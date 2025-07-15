'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Loader2, Wand2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Source } from '@doota/pb/doota/core/v1/core_pb'
import { useClientsContext } from '@doota/ui-core/context/ClientContext'
import { PostInsight } from '@doota/pb/doota/core/v1/insight_pb'
import { useCreatePost } from '@/components/hooks/useCreatePost'
import { PostSettings } from '@doota/pb/doota/core/v1/post_pb'

const toneOptions = [
    { value: 'professional', label: 'Professional' },
    { value: 'casual', label: 'Casual' },
    { value: 'friendly', label: 'Friendly' }
]

const goalOptions = [
    { value: 'karma', label: 'Build Karma' },
    { value: 'feedback', label: 'Get Feedback' },
    { value: 'leads', label: 'Generate Leads' }
]

export default function CreatePost() {
    const { portalClient } = useClientsContext()
    const { createPost } = useCreatePost()

    const [isLoading, setIsLoading] = useState(false)
    const [isPostApiCall, setIsPostApiCall] = useState(false)
    const [insights, setInsights] = useState<PostInsight[]>([])
    const [sources, setSources] = useState<Source[]>([])

    const [selectedInsight, setSelectedInsight] = useState('')
    const [selectedSubreddit, setSelectedSubreddit] = useState('')
    const [customTopic, setCustomTopic] = useState('')
    const [postDetails, setPostDetails] = useState('')
    const [selectedGoal, setSelectedGoal] = useState('')
    const [selectedTone, setSelectedTone] = useState('')

    useEffect(() => {
        setIsLoading(true)
        Promise.all([
            portalClient.getInsights({}).then(res => setInsights(res.insights)),
            portalClient.getSources({}).then(res => setSources(res.sources))
        ])
            .catch(err => console.error('Error fetching data:', err))
            .finally(() => setIsLoading(false))
    }, [])

    const handleGeneratePost = async () => {
        const postData: Omit<PostSettings, '$typeName'> = {
            referenceId: selectedInsight,
            sourceId: selectedSubreddit,
            topic: customTopic,
            context: postDetails,
            goal: selectedGoal,
            tone: selectedTone
        }
        setIsPostApiCall(true)
        await createPost(postData)
        setIsPostApiCall(false)
    }

    return (
            <div className='p-6 ml-[10%] mr-[10%]'>
                <h1 className='text-2xl font-bold mb-1'>Create New Post</h1>
                <p className='text-gray-500 mb-6'>
                    Generate AI-powered Reddit posts from insights or custom topics
                </p>

                <Card>
                    <CardContent className='p-6 space-y-6'>
                        {/* Insight Suggestions */}
                        <div>
                            <Label className='mb-1 block'>
                                Suggested Topics from Insights (Optional)
                            </Label>
                            <Select onValueChange={setSelectedInsight} value={selectedInsight}>
                                <SelectTrigger>
                                    <SelectValue placeholder='Select a suggested topic or leave blank to add your own...' />
                                </SelectTrigger>
                                <SelectContent>
                                    {insights.map(insight => (
                                        <SelectItem key={insight.id} value={insight.id}>
                                            <div className='flex items-center gap-2'>
                                                <Badge variant='secondary' className='text-xs'>
                                                    {insight.relevancyScore}%
                                                </Badge>
                                                <span className='truncate max-w-[300px] text-sm'>
                        {insight.topic}
                      </span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        {/* Topic */}
                        <div>
                            <Label htmlFor='topic'>Topic</Label>
                            <Textarea
                                id='topic'
                                value={customTopic}
                                onChange={e => setCustomTopic(e.target.value)}
                                placeholder='Enter your topic...'
                                className='min-h-[100px] text-base'
                            />
                        </div>

                        {/* Post Details */}
                        <div>
                            <Label htmlFor='details'>Post Details & Context</Label>
                            <Textarea
                                id='details'
                                value={postDetails}
                                onChange={e => setPostDetails(e.target.value)}
                                placeholder='Add specific details, context, examples, or requirements for your post...'
                                className='min-h-[150px] text-base'
                            />
                        </div>

                        {/* Options */}
                        <div className='grid grid-cols-1 sm:grid-cols-3 gap-4'>
                            <div>
                                <Label className='mb-1 block'>Target Subreddit</Label>
                                <Select value={selectedSubreddit} onValueChange={setSelectedSubreddit}>
                                    <SelectTrigger>
                                        <SelectValue placeholder='Select subreddit' />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {sources.map(source => (
                                            <SelectItem key={source.id} value={source.id}>
                                                {source.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <div>
                                <Label className='mb-1 block'>Post Goal</Label>
                                <Select value={selectedGoal} onValueChange={setSelectedGoal}>
                                    <SelectTrigger>
                                        <SelectValue placeholder='Select goal' />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {goalOptions.map(goal => (
                                            <SelectItem key={goal.value} value={goal.value}>
                                                {goal.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <div>
                                <Label className='mb-1 block'>Tone</Label>
                                <Select value={selectedTone} onValueChange={setSelectedTone}>
                                    <SelectTrigger>
                                        <SelectValue placeholder='Select tone' />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {toneOptions.map(tone => (
                                            <SelectItem key={tone.value} value={tone.value}>
                                                {tone.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        {/* Submit */}
                        <div className='flex justify-center pt-4'>
                            <Button
                                onClick={handleGeneratePost}
                                disabled={
                                    !customTopic ||
                                    !selectedSubreddit ||
                                    !selectedGoal ||
                                    !selectedTone ||
                                    !postDetails ||
                                    isPostApiCall
                                }
                                className='px-8 text-base'
                            >
                                <Wand2 className='h-4 w-4 mr-2' />
                                Generate Post with AI
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            </div>
    )
}
