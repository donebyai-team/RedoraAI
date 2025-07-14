'use client'

import { Post, PostSettings } from "@doota/pb/doota/core/v1/post_pb";
import { useRouter } from "next/navigation";
import { useAppDispatch } from "@/store/hooks";
import toast from "react-hot-toast";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { setPost } from "@/store/PostCreation/PostCreationSlice";
import {routes} from "@doota/ui-core/routing";

/**
 * Hook to create a new post and update Redux state.
 * @returns { createPost } - function to create post
 */
export function useCreatePost() {
    const router = useRouter();
    const dispatch = useAppDispatch();
    const { portalClient } = useClientsContext();

    const createPost = async (
        postData: Omit<PostSettings, "$typeName">,
        isCreateNewPost: boolean = true,
    ): Promise<Post | undefined> => { try {
            const res = await portalClient.createPost(postData);
            dispatch(setPost(res));
            if(!isCreateNewPost)
                toast.success('Post regenerated successfully!');

            if (isCreateNewPost) {
                router.push(routes.new.postCreationHub.editor);
            }

            return res;
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
            return undefined;
        }
    };

    return { createPost };
}
