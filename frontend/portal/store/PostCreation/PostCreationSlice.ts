import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import {
    Post,
} from "@doota/pb/doota/core/v1/post_pb";

export type PostState = Omit<Post, "$typeName">;

// Initial state with clean, empty Post values
const initialState: { post: PostState } = {
    post: {
        id: "",
        projectId: "",
        source: "",
        topic: "",
        description: "",
        status: "",
        reason: "",
    },
};

// Create the slice
const postSlice = createSlice({
    name: "postCreation",
    initialState,
    reducers: {
        // Accept full Post (with $typeName) and strip it for state
        setPost: (state, action: PayloadAction<Post | null>) => {
            if (action.payload) {
                const { ...rest } = action.payload;
                state.post = rest;
            }
        },
    },
});

// Export actions
export const {
    setPost,
} = postSlice.actions;

// Export reducer
export const postCreationReducer =  postSlice.reducer;
