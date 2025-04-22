"use client"

import { useState } from "react"
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    IconButton,
    Typography,
    Box,
} from "@mui/material"
import CloseIcon from "@mui/icons-material/Close"

interface AddSubredditDialogProps {
    open: boolean
    onClose: () => void
    onAdd: (subreddit: string) => void
}

export default function AddSubredditDialog({ open, onClose, onAdd }: AddSubredditDialogProps) {
    const [subreddit, setSubreddit] = useState("");

    const handleAdd = () => {
        if (subreddit.trim()) {
            onAdd(subreddit)
            setSubreddit("")
            onClose()
        }
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth="xs"
            PaperProps={{
                sx: {
                    borderRadius: 2,
                    padding: 6,
                    margin: 0
                },
            }}
        >
            <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", position: "relative" }}>
                <DialogTitle sx={{ fontWeight: "bold", fontSize: "1.25rem", p: 0, mb: 2.5 }}>Add subreddit to track</DialogTitle>
                <IconButton onClick={onClose} sx={{ position: "absolute", top: "-10px", right: "-10px" }} aria-label="close">
                    <CloseIcon />
                </IconButton>
            </Box>

            <DialogContent sx={{ p: 0 }}>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
                    Add URL or name of the subreddit you want to track.
                </Typography>
                <TextField
                    fullWidth
                    placeholder="Example: r/marketing"
                    value={subreddit}
                    size="small"
                    onChange={(e) => setSubreddit(e.target.value)}
                    variant="outlined"
                    InputProps={{
                        sx: {
                            borderRadius: 1.5,
                        },
                    }}
                />
            </DialogContent>

            <DialogActions sx={{ justifyContent: "flex-end", gap: 1, px: 0, pt: 4, pb: 0 }}>
                <Button
                    onClick={onClose}
                    variant="outlined"
                    sx={{
                        color: "text.primary",
                        textTransform: "none",
                        fontWeight: 500,
                        px: 3,
                    }}
                >
                    Cancel
                </Button>
                <Button
                    onClick={handleAdd}
                    variant="contained"
                    sx={{
                        bgcolor: "#f56f36",
                        "&:hover": {
                            bgcolor: "#e05f2a",
                        },
                        textTransform: "none",
                        fontWeight: 500,
                        borderRadius: 1,
                        px: 3,
                    }}
                >
                    Add subreddit
                </Button>
            </DialogActions>
        </Dialog>
    );
}
