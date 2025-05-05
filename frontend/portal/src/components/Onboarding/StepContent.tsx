import React, { useState } from 'react';
import {
    Box,
    Typography,
    TextField,
    Stack,
    Chip,
    Autocomplete,
    InputAdornment,
    IconButton,
    Paper,
} from '@mui/material';
import { Search, Plus, X } from 'lucide-react';

interface StepContentProps {
    step: number;
}

const StepContent: React.FC<StepContentProps> = ({ step }) => {
    const [keywords, setKeywords] = useState<string[]>([]);
    const [newKeyword, setNewKeyword] = useState('');
    const [selectedSubreddits, setSelectedSubreddits] = useState<string[]>([]);

    const StepCounter = () => (
        <Typography
            variant="body2"
            sx={{
                color: 'primary.main',
                fontWeight: 500,
                mb: 2,
                display: 'block'
            }}
        >
            Step {step + 1} of 3
        </Typography>
    );

    const handleAddKeyword = () => {
        if (newKeyword && !keywords.includes(newKeyword)) {
            setKeywords([...keywords, newKeyword]);
            setNewKeyword('');
        }
    };

    const handleDeleteKeyword = (keywordToDelete: string) => {
        setKeywords(keywords.filter(keyword => keyword !== keywordToDelete));
    };

    const popularSubreddits = [
        'technology', 'programming', 'webdev', 'startup',
        'business', 'marketing', 'entrepreneur', 'software'
    ];

    switch (step) {
        case 0:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Product Information
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Tell us about your product to help us track relevant discussions.
                    </Typography>

                    <Stack spacing={3}>
                        <TextField
                            fullWidth
                            label="Product Name"
                            placeholder="e.g., My Awesome Product"
                            variant="outlined"
                        />

                        <TextField
                            fullWidth
                            multiline
                            rows={3}
                            label="Product Description"
                            placeholder="Describe your product and its key features..."
                            variant="outlined"
                        />

                        <TextField
                            fullWidth
                            label="Product Website"
                            placeholder="https://example.com"
                            variant="outlined"
                        />

                        <TextField
                            fullWidth
                            label="Target Audience"
                            placeholder="e.g., Developers, Marketers, Small Business Owners"
                            variant="outlined"
                        />
                    </Stack>
                </Box>
            );

        case 1:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Track Keywords
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Add keywords related to your product that you want to track on Reddit.
                    </Typography>

                    <Stack spacing={3}>
                        <TextField
                            fullWidth
                            label="Add Keyword"
                            value={newKeyword}
                            onChange={(e) => setNewKeyword(e.target.value)}
                            onKeyPress={(e) => e.key === 'Enter' && handleAddKeyword()}
                            InputProps={{
                                endAdornment: (
                                    <InputAdornment position="end">
                                        <IconButton onClick={handleAddKeyword} edge="end">
                                            <Plus size={20} />
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            }}
                        />

                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                            {keywords.map((keyword) => (
                                <Chip
                                    key={keyword}
                                    label={keyword}
                                    onDelete={() => handleDeleteKeyword(keyword)}
                                    color="primary"
                                    variant="outlined"
                                />
                            ))}
                        </Box>

                        <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                            <Typography variant="subtitle2" gutterBottom>
                                Suggested Keywords
                            </Typography>
                            <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                                {['competitor', 'alternative', 'vs', 'review', 'help'].map((suggestion, index) => (
                                    <Chip
                                        key={index}
                                        label={suggestion}
                                        onClick={() => setNewKeyword(suggestion)}
                                        size="small"
                                    />
                                ))}
                            </Stack>
                        </Paper>
                    </Stack>
                </Box>
            );

        case 2:
            return (
                <Box>
                    <StepCounter />
                    <Typography variant="h5" gutterBottom color="primary">
                        Select Subreddits
                    </Typography>
                    <Typography color="text.secondary" sx={{ mb: 4 }}>
                        Choose subreddits where you want to track your keywords.
                    </Typography>

                    <Stack spacing={3}>
                        <Autocomplete
                            multiple
                            options={popularSubreddits}
                            value={selectedSubreddits}
                            onChange={(_, newValue) => setSelectedSubreddits(newValue)}
                            renderInput={(params) => (
                                <TextField
                                    {...params}
                                    label="Search Subreddits"
                                    placeholder="Type to search..."
                                    InputProps={{
                                        ...params.InputProps,
                                        startAdornment: (
                                            <>
                                                <InputAdornment position="start">
                                                    <Search size={20} />
                                                </InputAdornment>
                                                {params.InputProps.startAdornment}
                                            </>
                                        ),
                                    }}
                                />
                            )}
                            renderTags={(value, getTagProps) =>
                                value.map((option, index) => (
                                    <React.Fragment key={index}>
                                        <Chip
                                            label={option}
                                            {...getTagProps({ index })}
                                            color="primary"
                                            onDelete={() => {
                                                const newSelected = selectedSubreddits.filter(
                                                    (item) => item !== option
                                                );
                                                setSelectedSubreddits(newSelected);
                                            }}
                                            deleteIcon={<X size={16} />}
                                        />
                                    </React.Fragment>
                                ))
                            }
                        />

                        <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                            <Typography variant="subtitle2" gutterBottom>
                                Popular Subreddits
                            </Typography>
                            <Stack direction="row" spacing={1} flexWrap="wrap" gap={1}>
                                {popularSubreddits.slice(0, 6).map((subreddit) => (
                                    <Chip
                                        key={subreddit}
                                        label={`r/${subreddit}`}
                                        onClick={() => {
                                            if (!selectedSubreddits.includes(subreddit)) {
                                                setSelectedSubreddits([...selectedSubreddits, subreddit]);
                                            }
                                        }}
                                        size="small"
                                    />
                                ))}
                            </Stack>
                        </Paper>
                    </Stack>
                </Box>
            );

        default:
            return null;
    }
};

export default StepContent;