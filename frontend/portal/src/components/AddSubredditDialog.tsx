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
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const { portalClient } = useClientsContext()

    const validateSubreddit = (name: string) => {
        const trimmed = name.trim()
        const subredditRegex = /^(r\/[a-zA-Z0-9_]+|https?:\/\/(www\.)?reddit\.com\/r\/[a-zA-Z0-9_]+)/i
        return subredditRegex.test(trimmed)
    }

    const handleAdd = async () => {
        const trimmedSubreddit = subreddit.trim()
        setError(null)

        // Frontend validations
        if (!trimmedSubreddit) {
            setError("Subreddit is required.")
            return
        }

        if (!validateSubreddit(trimmedSubreddit)) {
            setError("Enter a valid subreddit (e.g., r/marketing or full Reddit URL).")
            return
        }

        setIsLoading(true)

        try {
            await portalClient.addSource({ name: trimmedSubreddit })
            onAdd(trimmedSubreddit)
            setSubreddit("")
            onClose()
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong"
            toast.error(message)
        } finally {
            setIsLoading(false)
        }
    }

    const handleDialogClose = () => {
        if (isLoading) return;
        setSubreddit("")
        setError(null)
        onClose()
    }

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
                    margin: 0,
                },
            }}
        >
            <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", position: "relative" }}>
                <DialogTitle sx={{ fontWeight: "bold", fontSize: "1.25rem", p: 0, mb: 2.5 }}>
                    Add subreddit to track
                </DialogTitle>
                <IconButton onClick={handleDialogClose} sx={{ position: "absolute", top: "-10px", right: "-10px" }} aria-label="close">
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
                    onChange={(e) => setSubreddit(e.target.value)}
                    size="small"
                    variant="outlined"
                    error={Boolean(error)}
                    helperText={error}
                    InputProps={{
                        sx: {
                            borderRadius: 1.5,
                        },
                    }}
                />
            </DialogContent>

            <DialogActions sx={{ justifyContent: "flex-end", gap: 1, px: 0, pt: 4, pb: 0 }}>
                <Button
                    onClick={handleDialogClose}
                    variant="outlined"
                    disabled={isLoading}
                    sx={{
                        color: "text.primary",
                        textTransform: "none",
                        fontWeight: 500,
                        px: 3,
                        opacity: isLoading ? 0.5 : 1,
                    }}
                >
                    Cancel
                </Button>
                <Button
                    onClick={handleAdd}
                    variant="contained"
                    disabled={isLoading}
                    sx={{
                        bgcolor: "#f56f36",
                        "&:hover": {
                            bgcolor: "#e05f2a",
                        },
                        textTransform: "none",
                        fontWeight: 500,
                        borderRadius: 1,
                        px: 3,
                        opacity: isLoading ? 0.5 : 1,
                    }}
                >
                    Add subreddit
                </Button>
            </DialogActions>
        </Dialog>
    )
}