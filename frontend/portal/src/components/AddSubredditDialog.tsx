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
import { useClientsContext } from "@doota/ui-core/context/ClientContext"
import toast from "react-hot-toast"

interface AddSubredditDialogProps {
    open: boolean
    onClose: () => void
    onAdd: (subreddit: string) => void
}

export default function AddSubredditDialog({ open, onClose, onAdd }: AddSubredditDialogProps) {
    const [subreddit, setSubreddit] = useState("")
    const { portalClient } = useClientsContext()

    const validateSubreddit = (name: string) => {
        const trimmed = name.trim()
        const subredditRegex = /^(r\/[a-zA-Z0-9_]+|https?:\/\/(www\.)?reddit\.com\/r\/[a-zA-Z0-9_]+)/i
        return subredditRegex.test(trimmed)
    }

    const handleAdd = async () => {
        const trimmedSubreddit = subreddit.trim()

        if (!trimmedSubreddit) {
            toast.error("Please enter a subreddit name.")
            return;
        }

        if (!validateSubreddit(trimmedSubreddit)) {
            toast.error("Enter a valid subreddit (e.g., r/marketing or full Reddit URL).")
            return;
        }

        const loadingToast = toast.loading("Adding subreddit...")

        try {
            await portalClient.addSubReddit({ name: trimmedSubreddit })
            toast.success("Subreddit added successfully.", { id: loadingToast })
            onAdd(trimmedSubreddit)
            setSubreddit("")
            onClose()
        } catch (error) {
            console.error("###_err", error)
            toast.error("Something went wrong. Please try again.", { id: loadingToast })
        }
    }

    const handleDialogClose = () => {
        setSubreddit("")
        onClose()
    };

    return (
        <Dialog
            open={open}
            onClose={handleDialogClose}
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
                    Add the URL or name of the subreddit you want to track.
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
