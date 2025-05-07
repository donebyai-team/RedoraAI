import React, { useState } from 'react';
import {
    Box,
    Typography,
    TextField,
    Stack,
    Chip,
    Autocomplete,
    InputAdornment,
    Paper,
} from '@mui/material';
import { Search, X } from 'lucide-react';
import ConnectRedditStep from './ConnectRedditStep';
import ProductInformationStep from './ProductInformationStep';
import TrackKeywordStep from './TrackKeywordStep';

interface StepContentProps {
    step: number;
    stepLength?: number
}

const StepContent: React.FC<StepContentProps> = ({ step, stepLength }) => {
    
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
            Step {step + 1} of {stepLength}
        </Typography>
    );

    

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

                    <ProductInformationStep />
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

                    <TrackKeywordStep />
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

        case 3:
            return <ConnectRedditStep />;

        default:
            return null;
    }
};

export default StepContent;